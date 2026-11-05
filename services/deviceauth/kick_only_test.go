package deviceauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestKickOnlyHistoricalExpiredJob(t *testing.T) {
	kicks := 0
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apigateway/users/kick":
			kicks++
			b, _ := io.ReadAll(r.Body)
			if string(b) != `{"user_id":"u","device_ids":["phone"]}` {
				t.Errorf("kick scope: %s", b)
			}
			io.WriteString(w, `{"code":0}`)
		case "/apigateway/users/onlinestatus/query":
			if kicks == 0 {
				io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone","platform":"iOS"}]}]}}`)
			} else {
				io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":false,"devices":[]}]}}`)
			}
		default:
			t.Errorf("forbidden IM endpoint: %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	j := RevokeJob{AppKey: "app", UserID: "u", DeviceID: "phone", OperationID: "op", TargetSessionID: "s", BannedUntil: 1}
	if err := RunIM(context.Background(), a, j); err != nil {
		t.Fatal(err)
	}
	if kicks != 1 {
		t.Fatal("kick count", kicks)
	}
}

func TestKickAcknowledgementDoesNotComplete(t *testing.T) {
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apigateway/users/kick" {
			io.WriteString(w, `{"code":0}`)
			return
		}
		if r.URL.Path != "/apigateway/users/onlinestatus/query" {
			t.Errorf("forbidden route %s", r.URL.Path)
		}
		io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone","platform":"iOS"}]}]}}`)
	})
	j := RevokeJob{AppKey: "app", UserID: "u", DeviceID: "phone", OperationID: "op", TargetSessionID: "s", BannedUntil: 999999}
	if err := RunIM(context.Background(), a, j); !errors.Is(err, ErrIM) {
		t.Fatal("async acknowledgement completed", err)
	}
}

func TestKickUnknownReplyReconcilesBeforeRetry(t *testing.T) {
	for _, path := range []string{"/apigateway/users/kick", "/apigateway/users/onlinestatus/query"} {
		t.Run(path, func(t *testing.T) {
			kicks := 0
			failed := false
			a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/apigateway/users/kick":
					kicks++
					if path == r.URL.Path {
						w.WriteHeader(503)
						return
					}
					io.WriteString(w, `{"code":0}`)
				case "/apigateway/users/onlinestatus/query":
					if kicks == 0 {
						io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone","platform":"iOS"}]}]}}`)
						return
					}
					if path == r.URL.Path && !failed {
						failed = true
						w.WriteHeader(503)
						return
					}
					io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":false,"devices":[]}]}}`)
				default:
					t.Errorf("forbidden route: %s", r.URL.Path)
					w.WriteHeader(500)
				}
			})
			j := RevokeJob{AppKey: "app", UserID: "u", DeviceID: "phone", OperationID: "op", TargetSessionID: "s"}
			if e := RunIM(context.Background(), a, j); !errors.Is(e, ErrIM) {
				t.Fatal("unknown result completed", e)
			}
			if e := RunIM(context.Background(), a, j); e != nil {
				t.Fatal(e)
			}
			if kicks != 1 {
				t.Fatal("blind kick retry", kicks)
			}
		})
	}
}
