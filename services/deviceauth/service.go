package deviceauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// Service credentials are opt-in until the external assistant's downstream
// dependencies are verified. They never use the user-session/legacy decoder.
type ServiceClaims struct {
	AppKey   string `json:"app_key"`
	UserID   string `json:"user_id"`
	Audience string `json:"aud"`
	Expires  int64  `json:"exp"`
	Issued   int64  `json:"iat"`
}

func SignService(app, user string, key []byte, now int64) (string, error) {
	if app == "" || user == "" || len(key) == 0 {
		return "", ErrAuth
	}
	b, e := json.Marshal(ServiceClaims{app, user, "jugglechat-assistant", now + 300000, now})
	if e != nil {
		return "", e
	}
	p := "svc1." + base64.RawURLEncoding.EncodeToString(b)
	m := hmac.New(sha256.New, key)
	m.Write([]byte(p))
	return p + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil)), nil
}
func VerifyService(token string, key []byte, app, path string, now int64) (ServiceClaims, error) {
	var c ServiceClaims
	p := strings.Split(token, ".")
	if len(p) != 3 || p[0] != "svc1" || len(key) == 0 {
		return c, ErrAuth
	}
	sig, e := base64.RawURLEncoding.DecodeString(p[2])
	if e != nil {
		return c, ErrAuth
	}
	m := hmac.New(sha256.New, key)
	m.Write([]byte(p[0] + "." + p[1]))
	if !hmac.Equal(sig, m.Sum(nil)) {
		return c, ErrAuth
	}
	b, e := base64.RawURLEncoding.DecodeString(p[1])
	if e != nil || json.Unmarshal(b, &c) != nil {
		return c, ErrAuth
	}
	if c.AppKey != app || c.UserID == "" || c.Audience != "jugglechat-assistant" || now >= c.Expires || now < c.Issued || c.Expires-c.Issued != 300000 {
		return c, ErrAuth
	}
	// Read-only local capabilities; extend only after downstream call inventory.
	switch path {
	case "/jim/users/info", "/jim/groups/info", "/jim/groups/members/list", "/jim/friends/list", "/jim/converconfs/get":
		return c, nil
	}
	return c, ErrAuth
}
