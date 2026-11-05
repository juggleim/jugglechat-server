package deviceauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrParams    = errors.New("invalid device parameters")
	ErrUpgrade   = errors.New("upgrade and sign in again with device parameters")
	ErrAuth      = errors.New("invalid or revoked device session")
	ErrNotFound  = errors.New("device not found")
	ErrCurrent   = errors.New("cannot delete current device")
	ErrIM        = errors.New("IM device kick unconfirmed or unavailable")
	ErrQR        = errors.New("invalid, expired or already claimed QR code")
	ErrQRExpired = errors.New("QR code expired")
	ErrPending   = errors.New("QR confirmation pending")
)

// RawMessage distinguishes omitted fields from explicit null and wrong JSON types.
type DeviceParams struct {
	DeviceID   json.RawMessage `json:"device_id"`
	Platform   json.RawMessage `json:"platform"`
	DeviceName string          `json:"device_name"`
	OSVersion  string          `json:"os_version"`
	AppVersion string          `json:"app_version"`
}

func validText(s string, n int) bool {
	if !utf8.ValidString(s) || utf8.RuneCountInString(s) > n {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// Reject boundary whitespace instead of normalizing identity. In particular,
// MySQL PAD SPACE collations can match trailing-space aliases that Go does not.
func validDeviceID(id string) bool {
	return id != "" && strings.TrimSpace(id) == id && validText(id, 128)
}
func (p DeviceParams) Identity() (string, string) {
	var d, s string
	json.Unmarshal(p.DeviceID, &d)
	json.Unmarshal(p.Platform, &s)
	return d, s
}
func (p DeviceParams) Validate(strict bool) (bool, error) {
	if !validText(p.DeviceName, 128) || !validText(p.OSVersion, 64) || !validText(p.AppVersion, 64) {
		return false, ErrParams
	}
	if len(p.DeviceID) == 0 && len(p.Platform) == 0 {
		if strict {
			return false, ErrUpgrade
		}
		return true, nil
	}
	d, s := p.Identity()
	if !validDeviceID(d) || !validText(s, 32) {
		return false, ErrParams
	}
	switch s {
	case "ios", "android", "web", "windows", "macos", "linux":
	default:
		return false, ErrParams
	}
	return false, nil
}

type Claims struct {
	AppKey    string `json:"app_key"`
	UserID    string `json:"user_id"`
	DeviceID  string `json:"device_id"`
	SessionID string `json:"session_id"`
	Kind      string `json:"kind"`
	IssuedAt  int64  `json:"issued_at"`
}

func (c Claims) valid() bool {
	return c.Kind == "user_session" && c.AppKey != "" && c.UserID != "" && c.DeviceID != "" && c.SessionID != ""
}
func Sign(c Claims, key []byte) (string, error) {
	if !c.valid() || len(key) == 0 {
		return "", ErrAuth
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	payload := "ds1." + base64.RawURLEncoding.EncodeToString(b)
	m := hmac.New(sha256.New, key)
	m.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil)), nil
}
func Verify(token string, key []byte, app string) (Claims, error) {
	var c Claims
	p := strings.Split(token, ".")
	if len(p) != 3 || p[0] != "ds1" || len(key) == 0 {
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
	if e != nil {
		return c, ErrAuth
	}
	if json.Unmarshal(b, &c) != nil || !c.valid() || c.AppKey != app {
		return Claims{}, ErrAuth
	}
	return c, nil
}
func RandomID() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type Device struct {
	ID                int64  `gorm:"primaryKey" json:"-"`
	AppKey            string `json:"-"`
	UserID            string `gorm:"column:user_id" json:"-"`
	DeviceID          string `gorm:"column:device_id" json:"device_id"`
	DeviceName        string `json:"device_name"`
	Platform          string `json:"platform"`
	OSVersion         string `gorm:"column:os_version" json:"os_version"`
	AppVersion        string `json:"app_version"`
	LastLoginTime     int64  `json:"last_login_time"`
	LastActiveTime    *int64 `json:"last_active_time"`
	SessionID         string `gorm:"column:session_id" json:"-"`
	State             string `json:"-"`
	RevokeOperationID string `gorm:"column:revoke_operation_id" json:"-"`
	BannedUntil       int64  `gorm:"->" json:"-"` // Deprecated history, never consulted or written.
	DeletedAt         *int64 `json:"-"`
	CreatedTime       int64  `json:"-"`
	UpdatedTime       int64  `json:"-"`
}

func (Device) TableName() string { return "user_devices" }

func (d Device) CanLogin(now int64) error {
	if d.State != "active" && d.State != "revoked" && d.State != "revoking" {
		return ErrAuth
	}
	return nil
}
func (d Device) Authenticate(c Claims, now int64) error {
	if !c.valid() || d.AppKey != c.AppKey || d.UserID != c.UserID || d.DeviceID != c.DeviceID || d.SessionID != c.SessionID || d.State != "active" || d.DeletedAt != nil {
		return ErrAuth
	}
	return nil
}
