package deviceauth

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/juggleim/jugglechat-server/storages/dbs"
)

func TestConverNotifySignedProtocol(t *testing.T) {
	for _, typ := range []int32{1, 2} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := "/apigateway/messages/private/send"
			if typ == 2 {
				path = "/apigateway/messages/group/send"
			}
			if r.URL.Path != path || r.Method != "POST" {
				t.Error(r.URL.Path)
			}
			sig := sha1.Sum([]byte("fixture" + r.Header.Get("nonce") + r.Header.Get("timestamp")))
			if r.Header.Get("appkey") != "a" || r.Header.Get("signature") != hex.EncodeToString(sig[:]) {
				t.Error("signature")
			}
			var body map[string]interface{}
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Error("body")
			}
			if body["sender_id"] != "alice" || body["target_id"] != "bob" || body["msg_type"] != "jgd:lifetime_msg" || body["msg_content"] != "{}" || body["is_storage"] != true || body["is_count"] != false || body["is_notify_sender"] != true {
				t.Error(body)
			}
			if body["is_cmd"] != true {
				t.Errorf("conver_type=%d: top-level is_cmd = %v, want true", typ, body["is_cmd"])
			}
			if _, ok := body["msg_id"]; ok {
				t.Error("invented idempotence")
			}
			w.Write([]byte(`{"code":0}`))
		}))
		adapter := &HTTPIM{Resolve: func(context.Context, string) (IMConfig, error) {
			return IMConfig{BaseURL: server.URL, AppKey: "a", Secret: "fixture"}, nil
		}}
		if e := adapter.NotifyConverConfig(context.Background(), dbs.ConverOutbox{AppKey: "a", ActorID: "alice", TargetID: "bob", ConverType: typ, Payload: "{}"}); e != nil {
			t.Fatal(e)
		}
		server.Close()
	}
}

func TestConverNotifyFailsClosedAndTimesOut(t *testing.T) {
	job := dbs.ConverOutbox{AppKey: "a", ActorID: "alice", TargetID: "bob", ConverType: 1, Payload: "{}"}
	if (&HTTPIM{}).NotifyConverConfig(context.Background(), job) == nil {
		t.Fatal("nil resolver")
	}
	for _, response := range []string{`{}`, `{"code":1,"msg":"secret upstream content"}`, `garbage`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(response)) }))
		a := &HTTPIM{Resolve: func(context.Context, string) (IMConfig, error) {
			return IMConfig{BaseURL: server.URL, AppKey: "a", Secret: "fixture"}, nil
		}}
		if e := a.NotifyConverConfig(context.Background(), job); e == nil {
			t.Fatal(response)
		}
		server.Close()
	}
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer server.Close()
	defer close(release)
	a := &HTTPIM{Timeout: 20 * time.Millisecond, Resolve: func(context.Context, string) (IMConfig, error) {
		return IMConfig{BaseURL: server.URL, AppKey: "a", Secret: "fixture"}, nil
	}}
	start := time.Now()
	if e := a.NotifyConverConfig(context.Background(), job); e == nil {
		t.Fatal("timeout success")
	}
	if time.Since(start) > time.Second {
		t.Fatal("unbounded")
	}
}
