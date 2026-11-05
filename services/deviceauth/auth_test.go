package deviceauth

import (
	"encoding/json"
	"errors"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"strings"
	"testing"
)

func TestDeviceIDRejectsBoundaryWhitespace(t *testing.T) {
	for _, id := range []string{" device-A", "device-A ", "\tdevice-A", "device-A\n", "\u00a0device-A", "device-A\u3000"} {
		for _, strict := range []bool{false, true} {
			p := DeviceParams{Platform: json.RawMessage(`"ios"`)}
			p.DeviceID, _ = json.Marshal(id)
			legacy, err := p.Validate(strict)
			if legacy || !errors.Is(err, ErrParams) {
				t.Errorf("id=%q strict=%v: legacy=%v err=%v", id, strict, legacy, err)
			}
			got, _ := p.Identity()
			if got != id {
				t.Fatalf("identity silently normalized: %q => %q", id, got)
			}
		}
	}
	p := DeviceParams{DeviceID: json.RawMessage(`"device A"`), Platform: json.RawMessage(`"ios"`)}
	if _, err := p.Validate(false); err != nil {
		t.Fatal("interior space rejected", err)
	}
}

func TestDeviceParametersCannotDowngrade(t *testing.T) {
	for _, tc := range []struct {
		body                string
		strict, legacy, bad bool
	}{
		{`{}`, false, true, false}, {`{}`, true, false, true},
		{`{"device_id":null}`, false, false, true}, {`{"device_id":""}`, false, false, true},
		{`{"device_id":"d"}`, false, false, true}, {`{"platform":"ios"}`, false, false, true},
		{`{"device_id":"d","platform":"bad"}`, false, false, true},
		{`{"device_id":"d","platform":"ios","device_name":"a\nb"}`, false, false, true},
		{`{"device_id":"d","platform":"ios"}`, true, false, false},
	} {
		var p DeviceParams
		if err := json.Unmarshal([]byte(tc.body), &p); err != nil {
			t.Fatal(err)
		}
		legacy, err := p.Validate(tc.strict)
		if legacy != tc.legacy || (err != nil) != tc.bad {
			t.Errorf("%s: legacy=%v err=%v", tc.body, legacy, err)
		}
	}
}

func TestStrictCutoffEquality(t *testing.T) {
	old := configures.Config.DeviceAuthRequiredAt
	defer func() { configures.Config.DeviceAuthRequiredAt = old }()
	configures.Config.DeviceAuthRequiredAt = map[string]int64{"a": 100}
	if Required("a", 99) || !Required("a", 100) || !Required("a", 101) || Required("other", 101) {
		t.Fatal("cutoff inconsistent")
	}
}

func TestTypedCredentialIsolation(t *testing.T) {
	c := Claims{AppKey: "app", UserID: "u", DeviceID: "d", SessionID: "s", Kind: "user_session", IssuedAt: 100}
	token, err := Sign(c, []byte("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verify(token, []byte("test-key"), "app")
	if err != nil || got != c {
		t.Fatalf("roundtrip: %v %v", got, err)
	}
	for _, tc := range []struct{ token, app string }{{token, "other"}, {token + "x", "app"}, {strings.Replace(token, "ds1.", "ds2.", 1), "app"}} {
		if _, err := Verify(tc.token, []byte("test-key"), tc.app); err == nil {
			t.Fatal("invalid credential accepted")
		}
	}
	c.SessionID = ""
	if _, err := Sign(c, []byte("test-key")); err == nil {
		t.Fatal("partial credential accepted")
	}
}

func TestSessionRevocationAndRotation(t *testing.T) {
	d := Device{AppKey: "app", UserID: "u", DeviceID: "d", SessionID: "new", State: "active"}
	c := Claims{AppKey: "app", UserID: "u", DeviceID: "d", SessionID: "old", Kind: "user_session"}
	if d.Authenticate(c, 100) == nil {
		t.Fatal("rotated token accepted")
	}
	c.SessionID = "new"
	if err := d.Authenticate(c, 100); err != nil {
		t.Fatal(err)
	}
	d.State = "revoking"
	if d.Authenticate(c, 100) == nil {
		t.Fatal("revoked token accepted")
	}
	if err := d.CanLogin(1000000); err != nil {
		t.Fatal("pending revocation blocked login", err)
	}
	d.State = "revoked"
	d.BannedUntil = 200
	if err := d.CanLogin(199); err != nil {
		t.Fatal("deprecated ban blocked login", err)
	}
	if err := d.CanLogin(200); err != nil {
		t.Fatal(err)
	}
}
