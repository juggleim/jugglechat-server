package deviceauth

import (
	"context"
	"errors"
	"testing"
)

func TestRevocationRetryWithoutBan(t *testing.T) {
	d := Device{AppKey: "a", UserID: "u", DeviceID: "d", SessionID: "s", State: "active"}
	j, e := d.BeginRevoke("op", 100)
	if e != nil {
		t.Fatal(e)
	}
	if d.State != "revoking" || d.BannedUntil != 0 || j.BannedUntil != 0 || j.TargetSessionID != "s" {
		t.Fatalf("%+v %+v", d, j)
	}
	again, e := d.BeginRevoke("other", 200)
	if e != nil || again.OperationID != "op" || again.BannedUntil != 0 {
		t.Fatalf("%+v %v", again, e)
	}
	if e := RunIM(context.Background(), UnsupportedIM{}, j); !errors.Is(e, ErrIM) {
		t.Fatal(e)
	}
	if d.State != "revoking" {
		t.Fatal("failed IM marked complete")
	}
}

func TestRevokingAndHistoricalBanAllowLogin(t *testing.T) {
	for _, state := range []string{"active", "revoking", "revoked"} {
		d := Device{State: state, BannedUntil: 999999}
		if err := d.CanLogin(100); err != nil {
			t.Errorf("%s: %v", state, err)
		}
	}
	d := Device{AppKey: "a", UserID: "u", DeviceID: "d", SessionID: "new", State: "active", BannedUntil: 999999}
	c := Claims{"a", "u", "d", "new", "user_session", 100}
	if err := d.Authenticate(c, 101); err != nil {
		t.Fatal(err)
	}
	c.SessionID = "old"
	if err := d.Authenticate(c, 9999999); !errors.Is(err, ErrAuth) {
		t.Fatal("old session revived", err)
	}
}
