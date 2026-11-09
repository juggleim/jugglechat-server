package services

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestPasswordSMSConfigFallback(t *testing.T) {
	const valid = `{"channel":"aliyun","aliyun":{"access_key":"test-key","access_secret":"test-secret","sign_name":"verification","template_code":"test-template"}}`
	for _, tc := range []struct {
		name, explicit       string
		err                  error
		wantOK, wantFallback bool
	}{
		{"missing falls back", "", gorm.ErrRecordNotFound, true, true},
		{"explicit valid wins", valid, nil, true, false},
		{"empty is invalid", "", nil, false, false},
		{"malformed is invalid", "{", nil, false, false},
		{"incomplete is invalid", `{"channel":"aliyun"}`, nil, false, false},
		{"database failure fails closed", "", errors.New("database unavailable"), false, false},
		{"unsupported provider fails closed", `{"channel":"smsbao","smsbao":{"username":"test","password":"test","template":"code"}}`, nil, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fallback := false
			engine, err := resolvePasswordSMSEngine(func(key string) (string, error) {
				switch key {
				case "set_password_sms_engine_conf":
					return tc.explicit, tc.err
				case "sms_engine_conf":
					fallback = true
					return valid, nil
				default:
					t.Fatal("unexpected configuration key")
					return "", errors.New("unexpected key")
				}
			})
			if (err == nil && engine != nil) != tc.wantOK || fallback != tc.wantFallback {
				t.Fatalf("success=%v fallback=%v", err == nil && engine != nil, fallback)
			}
		})
	}
	for _, tc := range []struct {
		value string
		err   error
	}{
		{"", gorm.ErrRecordNotFound}, {"", errors.New("database failure")},
		{`{"channel":"baidu"}`, nil}, {"{", nil},
	} {
		engine, err := resolvePasswordSMSEngine(func(key string) (string, error) {
			if key == "set_password_sms_engine_conf" {
				return "", gorm.ErrRecordNotFound
			}
			return tc.value, tc.err
		})
		if err == nil || engine != nil {
			t.Fatal("invalid fallback accepted")
		}
	}
}
