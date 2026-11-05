package deviceauth

import (
	"context"
	"errors"
	"testing"
)

func TestUnavailableIMNeverInventsOfflineOrSuccess(t *testing.T) {
	a := UnsupportedIM{}
	if _, e := a.Online(context.Background(), "a", "u"); !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if _, e := a.Reconcile(context.Background(), RevokeJob{}); !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if e := a.Disconnect(context.Background(), RevokeJob{}); !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
}
func TestGroupingStableAndScoped(t *testing.T) {
	ds := []Device{{AppKey: "a", UserID: "u", DeviceID: "z", State: "active", LastLoginTime: 2}, {AppKey: "a", UserID: "u", DeviceID: "b", State: "active", LastLoginTime: 3}, {AppKey: "a", UserID: "u", DeviceID: "a", State: "active", LastLoginTime: 3}, {AppKey: "other", UserID: "u", DeviceID: "x", State: "active"}, {AppKey: "a", UserID: "u", DeviceID: "gone", State: "revoking"}}
	g := Group(ds, map[string]bool{"b": true}, "a", "u", "a")
	if len(g.Online) != 1 || g.Online[0].DeviceID != "b" || len(g.Offline) != 2 || g.Offline[0].DeviceID != "a" || !g.Offline[0].IsCurrent {
		t.Fatalf("%+v", g)
	}
}
