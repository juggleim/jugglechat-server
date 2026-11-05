package deviceauth

import (
	"errors"
	"testing"
)

func TestQRPinsModeDeviceAndSecret(t *testing.T) {
	q := QRRecord{Status: "confirmed", SecretHash: SecretHash("secret"), CreatedTime: 100}
	if e := q.CheckClaim(false, "d", "wrong", 101); e == nil {
		t.Fatal("wrong secret accepted")
	}
	if e := q.CheckClaim(false, "d", "secret", 101); e != nil {
		t.Fatal(e)
	}
	q.Status = "consumed"
	q.Mode = "device"
	q.DeviceID = "d"
	for _, tc := range []struct {
		legacy         bool
		device, secret string
	}{{true, "", ""}, {false, "other", "secret"}, {false, "d", ""}} {
		if q.CheckClaim(tc.legacy, tc.device, tc.secret, 102) == nil {
			t.Fatal("claim switched")
		}
	}
	if e := q.CheckClaim(false, "d", "secret", 102); e != nil {
		t.Fatal(e)
	}
	if q.CheckClaim(false, "d", "secret", 600100) == nil {
		t.Fatal("expired accepted")
	}
	q.Mode = "legacy"
	if q.CheckClaim(false, "d", "secret", 102) == nil {
		t.Fatal("legacy switched to device")
	}
}

func TestQRExpiryRetainsLegacyProtocolError(t *testing.T) {
	q := QRRecord{Status: "confirmed", CreatedTime: 100, Mode: "legacy"}
	if e := q.CheckClaim(true, "", "", 600100); !errors.Is(e, ErrQRExpired) {
		t.Fatal(e)
	}
}
func TestQRResultEncryptedAndBound(t *testing.T) {
	cipher, e := SealResult([]byte("secret-key"), "app:qr", []byte("sensitive token"))
	if e != nil {
		t.Fatal(e)
	}
	plain, e := OpenResult([]byte("secret-key"), "app:qr", cipher)
	if e != nil || string(plain) != "sensitive token" {
		t.Fatal(e)
	}
	if _, e := OpenResult([]byte("secret-key"), "other:qr", cipher); e == nil {
		t.Fatal("result moved between QR records")
	}
	if _, e := OpenResult([]byte("wrong-key"), "app:qr", cipher); e == nil {
		t.Fatal("wrong encryption key")
	}
}
