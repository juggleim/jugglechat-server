package services

import (
	"github.com/juggleim/jugglechat-server/services/pbobjs"
	"testing"
)

func TestMalformedLegacyCiphertextRejectedWithoutPanic(t *testing.T) {
	for _, b := range [][]byte{nil, {1}, {1, 2, 3}} {
		if _, e := ParseToken(&pbobjs.AuthToken{Appkey: "a", TokenValue: b}, []byte("1234567890123456")); e == nil {
			t.Fatal("malformed token accepted")
		}
	}
}
func TestLegacyTokenRoundTripAndDeviceField(t *testing.T) {
	key := []byte("1234567890123456")
	original := ImToken{AppKey: "a", UserId: "u", TokenTime: 123}
	token, e := original.ToTokenString(key)
	if e != nil {
		t.Fatal(e)
	}
	wrap, e := ParseTokenString(token)
	if e != nil {
		t.Fatal(e)
	}
	got, e := ParseToken(wrap, key)
	if e != nil || got != original {
		t.Fatalf("%+v %v", got, e)
	}
}
