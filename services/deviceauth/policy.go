package deviceauth

import "strings"

func Allowed(kind, path string, strict bool) bool {
	sensitive := strings.HasPrefix(path, "/jim/devices/") || path == "/jim/login/qrcode/confirm"
	switch kind {
	case "user_session":
		return true
	case "legacy":
		return !strict && !strings.HasPrefix(path, "/jim/devices/")
	case "api_key":
		return !sensitive
	}
	return false
}
