package deviceauth

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testHTTPIM(t *testing.T, handler http.HandlerFunc) *HTTPIM {
	t.Helper()
	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)
	return &HTTPIM{Resolve: func(ctx context.Context, app string) (IMConfig, error) {
		return IMConfig{BaseURL: s.URL, AppKey: app, Secret: "test-secret", DeviceKick: true}, nil
	}, Timeout: 100 * time.Millisecond, Now: func() time.Time { return time.UnixMilli(1000) }}
}

func TestHTTPIMOnlineSignatureAndDevices(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		sum := sha1.Sum([]byte("test-secret" + r.Header.Get("nonce") + r.Header.Get("timestamp")))
		if r.Header.Get("signature") != hex.EncodeToString(sum[:]) || r.Header.Get("appkey") != "app" {
			t.Error("invalid signature")
		}
		b, _ := io.ReadAll(r.Body)
		if r.URL.Path != "/apigateway/users/onlinestatus/query" || string(b) != `{"user_ids":["u"]}` {
			t.Error("invalid request")
		}
		io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone","platform":"iOS"}]}]}}`)
	})
	got, e := a.Online(context.Background(), "app", "u")
	if e != nil || !got["phone"] || len(got) != 1 {
		t.Fatal(got, e)
	}
}

func TestHTTPIMOnlineRejectsIncompleteResponse(t *testing.T) {
	for _, body := range []string{`{}`, `{"code":12}`, `{"code":null}`, `{"code":0}`, `{"code":0,"data":null}`, `{"code":0,"data":{"items":[]}}`, `{"code":0,"data":{"items":[{"user_id":"other","is_online":false,"devices":[]}]}}`, `{"code":0,"data":{"items":[{"user_id":"u","is_online":false}]}}`, `{"code":0,"data":{"items":[{"user_id":"u","devices":[]}]}}`, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[]}]}}`, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[null]}]}}`, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"d","platform":"iOS"},{"device_id":"d","platform":"iOS"}]}]}}`} {
		t.Run(body, func(t *testing.T) {
			a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) })
			if _, e := a.Online(context.Background(), "app", "u"); !errors.Is(e, ErrIM) {
				t.Fatal(e)
			}
		})
	}
}

func TestHTTPIMOfflineCompletesWithoutAnyWrite(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apigateway/users/onlinestatus/query" {
			t.Errorf("forbidden route %s", r.URL.Path)
		}
		io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":false,"devices":[]}]}}`)
	})
	if e := RunIM(context.Background(), a, RevokeJob{AppKey: "app", UserID: "u", DeviceID: "phone", OperationID: "op", TargetSessionID: "s", BannedUntil: 1}); e != nil {
		t.Fatal(e)
	}
}

func TestHTTPIMDisabledAndInvalidNeverCallRemote(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) { t.Error("must not call remote") })
	j := RevokeJob{AppKey: "app", UserID: "u", DeviceID: " phone", OperationID: "op", TargetSessionID: "s"}
	if e := RunIM(context.Background(), a, j); !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	j.DeviceID = "phone"
	resolve := a.Resolve
	a.Resolve = func(ctx context.Context, app string) (IMConfig, error) {
		c, e := resolve(ctx, app)
		c.DeviceKick = false
		return c, e
	}
	if e := RunIM(context.Background(), a, j); !errors.Is(e, ErrIM) {
		t.Fatal("disabled silently succeeded", e)
	}
}

func TestHTTPIMTimeoutAndCancellation(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})
	a.Timeout = 20 * time.Millisecond
	if _, e := a.Online(context.Background(), "app", "u"); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := a.Online(ctx, "app", "u"); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}

func TestHTTPIMConfigurationAndSecretRedaction(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid configuration reached remote") })
	for _, c := range []IMConfig{
		{AppKey: "other", Secret: "secret"},
		{AppKey: "app", BaseURL: "https://example.invalid"},
		{AppKey: "app", Secret: "secret", BaseURL: "http://user:private@example.invalid"},
		{AppKey: "app", Secret: "secret", BaseURL: "https://example.invalid?private=value"},
	} {
		a.Resolve = func(context.Context, string) (IMConfig, error) { return c, nil }
		if _, e := a.Online(context.Background(), "app", "u"); !errors.Is(e, ErrIM) || strings.Contains(e.Error(), "private") || strings.Contains(e.Error(), "secret") {
			t.Fatal(e)
		}
	}
}
