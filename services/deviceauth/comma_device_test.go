package deviceauth

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestHTTPIMCommaDeviceReconcileAndKickExactArray(t *testing.T) {
	kicks := 0
	a := testHTTPIM(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/apigateway/users/kick":
			b, _ := io.ReadAll(r.Body)
			if string(b) != `{"user_id":"u","device_ids":["phone,other"]}` {
				t.Errorf("target split or widened: %s", b)
			}
			kicks++
			io.WriteString(w, `{"code":0}`)
		case "/apigateway/users/onlinestatus/query":
			if kicks == 0 {
				io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone,other","platform":"iOS"},{"device_id":"phone","platform":"iOS"}]}]}}`)
			} else {
				io.WriteString(w, `{"code":0,"data":{"items":[{"user_id":"u","is_online":true,"devices":[{"device_id":"phone","platform":"iOS"}]}]}}`)
			}
		default:
			t.Errorf("forbidden endpoint %s", r.URL.Path)
			w.WriteHeader(500)
		}
	})
	j := RevokeJob{AppKey: "app", UserID: "u", DeviceID: "phone,other", OperationID: "op", TargetSessionID: "s"}
	if done, err := a.Reconcile(context.Background(), j); err != nil || done {
		t.Fatalf("online exact target: done=%v err=%v", done, err)
	}
	if err := RunIM(context.Background(), a, j); err != nil {
		t.Fatal(err)
	}
	if done, err := a.Reconcile(context.Background(), j); err != nil || !done {
		t.Fatalf("offline exact target: done=%v err=%v", done, err)
	}
	if kicks != 1 {
		t.Fatal("kick count", kicks)
	}
}
