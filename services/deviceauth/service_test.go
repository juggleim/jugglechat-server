package deviceauth

import "testing"

func TestServiceCredentialExpiresAndCannotManageDevices(t *testing.T) {
	token, e := SignService("app", "u", []byte("key"), 100)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := VerifyService(token, []byte("key"), "app", "/jim/users/info", 101); e != nil {
		t.Fatal(e)
	}
	for _, path := range []string{"/jim/devices/list", "/jim/devices/del", "/jim/login/qrcode/confirm", "/jim/users/del"} {
		if _, e := VerifyService(token, []byte("key"), "app", path, 101); e == nil {
			t.Fatal(path)
		}
	}
	if _, e := VerifyService(token, []byte("key"), "other", "/jim/users/info", 101); e == nil {
		t.Fatal("wrong app")
	}
	if _, e := VerifyService(token, []byte("key"), "app", "/jim/users/info", 300100); e == nil {
		t.Fatal("expired")
	}
	if _, e := Verify(token, []byte("key"), "app"); e == nil {
		t.Fatal("service used as user")
	}
}
