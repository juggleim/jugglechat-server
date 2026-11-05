package deviceauth

import "testing"

func TestCredentialPolicy(t *testing.T) {
	for _, tc := range []struct {
		kind, path   string
		strict, want bool
	}{
		{"legacy", "/jim/users/info", false, true}, {"legacy", "/jim/users/info", true, false},
		{"legacy", "/jim/devices/list", false, false}, {"legacy", "/jim/login/qrcode/confirm", false, true}, {"legacy", "/jim/login/qrcode/confirm", true, false},
		{"api_key", "/jim/users/info", true, true}, {"api_key", "/jim/login/qrcode/confirm", false, false}, {"api_key", "/jim/devices/del", false, false},
		{"user_session", "/jim/devices/list", true, true}, {"service", "/jim/devices/list", false, false},
	} {
		if Allowed(tc.kind, tc.path, tc.strict) != tc.want {
			t.Errorf("%+v", tc)
		}
	}
}
