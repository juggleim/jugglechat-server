package deviceauth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QRRecord struct {
	CodeID       string `gorm:"column:code_id;primaryKey"`
	AppKey       string `gorm:"primaryKey"`
	UserID       string `gorm:"column:user_id"`
	Status       string
	Mode         string
	SecretHash   string
	DeviceID     string `gorm:"column:device_id"`
	SessionID    string `gorm:"column:session_id"`
	CreatedTime  int64
	ConsumedTime int64
	ResultCipher string
}

func (QRRecord) TableName() string { return "device_login_qrs" }
func SecretHash(s string) string   { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }
func (q QRRecord) CheckClaim(legacy bool, id, secret string, now int64) error {
	if now >= q.CreatedTime+600000 {
		return ErrQRExpired
	}
	if now < q.CreatedTime {
		return ErrQR
	}
	if legacy {
		if secret != "" || q.Mode == "device" {
			return ErrQR
		}
	} else {
		if secret == "" || !hmac.Equal([]byte(SecretHash(secret)), []byte(q.SecretHash)) || q.Mode == "legacy" {
			return ErrQR
		}
		if q.Status == "consumed" && q.DeviceID != id {
			return ErrQR
		}
	}
	if q.Status == "pending" {
		return ErrPending
	}
	if q.Status != "confirmed" && q.Status != "consumed" {
		return ErrQR
	}
	return nil
}
func resultAEAD(key []byte) (cipher.AEAD, error) {
	k := sha256.Sum256(append([]byte("device-qr-result:"), key...))
	b, e := aes.NewCipher(k[:])
	if e != nil {
		return nil, e
	}
	return cipher.NewGCM(b)
}
func SealResult(key []byte, aad string, data []byte) (string, error) {
	a, e := resultAEAD(key)
	if e != nil {
		return "", e
	}
	n := make([]byte, a.NonceSize())
	if _, e = rand.Read(n); e != nil {
		return "", e
	}
	return base64.RawStdEncoding.EncodeToString(a.Seal(n, n, data, []byte(aad))), nil
}
func OpenResult(key []byte, aad, s string) ([]byte, error) {
	a, e := resultAEAD(key)
	if e != nil {
		return nil, e
	}
	b, e := base64.RawStdEncoding.DecodeString(s)
	if e != nil || len(b) < a.NonceSize() {
		return nil, ErrQR
	}
	return a.Open(nil, b[:a.NonceSize()], b[a.NonceSize():], []byte(aad))
}
func (s Store) CreateQR(ctx context.Context, app, id string, now int64) (string, error) {
	secret, e := RandomID()
	if e != nil {
		return "", e
	}
	q := QRRecord{CodeID: id, AppKey: app, Status: "pending", SecretHash: SecretHash(secret), CreatedTime: now}
	return secret, s.DB.WithContext(ctx).Create(&q).Error
}
func (s Store) ConfirmQR(ctx context.Context, app, id, user string, now int64) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var q QRRecord
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_key=? AND code_id=?", app, id).Take(&q).Error; e != nil {
			return ErrQR
		}
		if now >= q.CreatedTime+600000 {
			return ErrQR
		}
		if q.Status != "pending" {
			if q.UserID == user {
				return nil
			}
			return ErrQR
		}
		return tx.Model(&q).Updates(map[string]interface{}{"status": "confirmed", "user_id": user}).Error
	})
}

// LoginResult is also the encrypted, ten-minute QR retry result. IM credentials
// are never kept in plaintext in QR rows or audit logs.
type LoginResult struct {
	UserID        string `json:"user_id"`
	Authorization string `json:"authorization"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Status        int    `json:"status"`
	IMToken       string `json:"im_token,omitempty"`
}

// PrepareQR does the network/user lookup outside the database transaction.
// RedeemQR validates again under the lock before issuing or releasing a result.
func (s Store) PrepareQR(ctx context.Context, app, id string, p DeviceParams, secret string, now int64) (QRRecord, error) {
	legacy, e := p.Validate(Required(app, now))
	if e != nil {
		return QRRecord{}, e
	}
	d, _ := p.Identity()
	var q QRRecord
	if e = s.DB.WithContext(ctx).Where("app_key=? AND code_id=?", app, id).Take(&q).Error; e != nil {
		return q, ErrQR
	}
	return q, q.CheckClaim(legacy, d, secret, now)
}
func (s Store) RedeemQR(ctx context.Context, app, id string, p DeviceParams, secret string, key []byte, now int64, prepared LoginResult, legacySign func(string) (string, error)) (LoginResult, error) {
	legacy, e := p.Validate(Required(app, now))
	if e != nil {
		return LoginResult{}, e
	}
	device, _ := p.Identity()
	var result LoginResult
	e = s.tx(ctx, func(tx *gorm.DB) error {
		var q QRRecord
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_key=? AND code_id=?", app, id).Take(&q).Error; e != nil {
			return ErrQR
		}
		if e := q.CheckClaim(legacy, device, secret, now); e != nil {
			return e
		}
		if q.Status == "consumed" {
			if !legacy {
				var d Device
				if e := deviceQuery(tx, app, q.UserID, device).Take(&d).Error; e != nil {
					return ErrAuth
				}
				if e := d.Authenticate(Claims{app, q.UserID, device, q.SessionID, "user_session", q.ConsumedTime}, now); e != nil {
					return e
				}
			}
			b, e := OpenResult(key, app+":"+id, q.ResultCipher)
			if e != nil {
				return e
			}
			return json.Unmarshal(b, &result)
		}
		if prepared.UserID != q.UserID {
			return ErrQR
		}
		result = prepared
		if legacy {
			token, e := legacySign(q.UserID)
			if e != nil {
				return e
			}
			result.Authorization = token
			q.Mode = "legacy"
		} else {
			d, e := issueTx(tx, app, q.UserID, p, now)
			if e != nil {
				return e
			}
			result.Authorization, e = Sign(Claims{app, q.UserID, device, d.SessionID, "user_session", now}, key)
			if e != nil {
				return e
			}
			q.Mode = "device"
			q.DeviceID = device
			q.SessionID = d.SessionID
		}
		b, e := json.Marshal(result)
		if e != nil {
			return e
		}
		q.ResultCipher, e = SealResult(key, app+":"+id, b)
		if e != nil {
			return e
		}
		q.Status = "consumed"
		q.ConsumedTime = now
		return tx.Save(&q).Error
	})
	if e == nil && legacy {
		LegacyQRRedemptions.Add(1)
	}
	return result, e
}
