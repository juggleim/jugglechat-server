package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/smsengines"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type passwordSMSService struct {
	Pepper  string
	DB      *gorm.DB
	Now     func() time.Time
	Send    func(context.Context, string, string, string) error
	Prepare func(context.Context, string) (smsengines.ISmsEngine, error)
}
type passwordSMSChallenge struct {
	ID                   string `gorm:"primaryKey"`
	Phone, Digest, Nonce string
	ExpiresAt            int64
	Failures             int
	Delivered            bool
	PhoneVersion         int64
}

func (passwordSMSChallenge) TableName() string { return "password_sms_challenges" }

type passwordSMSQuota struct {
	ID                            string `gorm:"primaryKey"`
	LastSend, HourStart, DayStart int64
	HourCount, DayCount           int
	AttemptNonce                  string
}

func (passwordSMSQuota) TableName() string { return "password_sms_quotas" }
func passwordSMSKey(parts ...string) string {
	b, _ := json.Marshal(parts)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (s passwordSMSService) digest(app, user, phone, nonce, code string, version int64) string {
	b, _ := json.Marshal([]string{"set_password", app, user, phone, nonce, code, strconv.FormatInt(version, 10)})
	mac := hmac.New(sha256.New, []byte(s.Pepper))
	mac.Write(b)
	return hex.EncodeToString(mac.Sum(nil))
}
func validInitialPassword(p string) bool {
	return utf8.ValidString(p) && utf8.RuneCountInString(p) >= 6 && utf8.RuneCountInString(p) <= 64 && strings.TrimSpace(p) != ""
}
func passwordSMSIdentity(ctx context.Context) (string, string, bool) {
	app := ctxs.GetAppKeyFromCtx(ctx)
	user := ctxs.GetRequesterIdFromCtx(ctx)
	kind, _ := ctx.Value(ctxs.CtxKey_AuthKind).(string)
	return app, user, app != "" && user != "" && (kind == "legacy" || kind == "user_session")
}

type passwordSMSUser struct {
	dbs.UserDao
	PasswordSMSPhoneVersion int64
}

func (s passwordSMSService) transaction(ctx context.Context, fn func(*gorm.DB, *passwordSMSUser) errs.IMErrorCode) errs.IMErrorCode {
	app, user, ok := passwordSMSIdentity(ctx)
	if !ok {
		return errs.IMErrorCode_APP_NOT_LOGIN
	}
	result := errs.IMErrorCode_SUCCESS
	err := s.DB.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		var u passwordSMSUser
		e := tx.Table("users").Select("id", "phone", "login_pass", "password_sms_phone_version").Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_key=? AND user_id=?", app, user).Take(&u).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			result = errs.IMErrorCode_APP_USER_NOT_EXIST
			return nil
		}
		if e != nil {
			return e
		}
		if u.LoginPass != "" {
			result = errs.IMErrorCode_APP_PASSWORD_ALREADY_SET
			return nil
		}
		if !passwordStatusPhone.MatchString(u.Phone) {
			result = errs.IMErrorCode_APP_PASSWORD_PHONE_REQUIRED
			return nil
		}
		if len(s.Pepper) < 32 {
			result = errs.IMErrorCode_APP_PASSWORD_SMS_DISABLED
			return nil
		}
		result = fn(tx, &u)
		if tx.Error != nil {
			return tx.Error
		}
		return nil
	})
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	return result
}

func (s passwordSMSService) send(ctx context.Context) errs.IMErrorCode {
	code, _ := s.sendWithRetry(ctx)
	return code
}

func (s passwordSMSService) sendWithRetry(ctx context.Context) (errs.IMErrorCode, int64) {
	var retry int64
	code := s.sendAttempt(ctx, &retry)
	return code, retry
}

// Reserve both quotas atomically. Definite rejection releases only this attempt's
// short cooldown; hourly/daily abuse budgets remain charged.
func (s passwordSMSService) sendAttempt(ctx context.Context, retry *int64) errs.IMErrorCode {
	if _, _, ok := passwordSMSIdentity(ctx); !ok {
		return errs.IMErrorCode_APP_NOT_LOGIN
	}
	if len(s.Pepper) < 32 {
		return errs.IMErrorCode_APP_PASSWORD_SMS_DISABLED
	}
	app, user, _ := passwordSMSIdentity(ctx)
	send := s.Send
	n, e := rand.Int(rand.Reader, big.NewInt(1000000))
	if e != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	code := fmt.Sprintf("%06d", n.Int64())
	// The legacy universal code is never issued or accepted by this flow.
	if code == "000000" {
		return s.sendAttempt(ctx, retry)
	}
	nonceBytes := make([]byte, 32)
	if _, e = rand.Read(nonceBytes); e != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	nonce := hex.EncodeToString(nonceBytes)
	id := passwordSMSKey("set_password", app, user)
	var phone string
	var quotaKeys []string
	result := s.transaction(ctx, func(tx *gorm.DB, u *passwordSMSUser) errs.IMErrorCode {
		if s.Prepare != nil {
			engine, err := s.Prepare(ctx, app)
			if err != nil {
				return errs.IMErrorCode_APP_SMS_SEND_FAILED
			}
			send = func(_ context.Context, _, phone, code string) error {
				return engine.SmsSend(phone, map[string]interface{}{"code": code})
			}
		}
		now := s.Now().Unix()
		phone = u.Phone
		quotas := []passwordSMSQuota{}
		quotaKeys = []string{passwordSMSKey("set_password", app, "account", user), passwordSMSKey("set_password", app, "phone", phone)}
		for _, key := range quotaKeys {
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&passwordSMSQuota{ID: key}).Error; e != nil {
				tx.AddError(e)
				return errs.IMErrorCode_APP_DEFAULT
			}
			var q passwordSMSQuota
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", key).Take(&q).Error; e != nil {
				tx.AddError(e)
				return errs.IMErrorCode_APP_DEFAULT
			}
			if now-q.HourStart >= 3600 {
				q.HourStart = now
				q.HourCount = 0
			}
			if now-q.DayStart >= 86400 {
				q.DayStart = now
				q.DayCount = 0
			}
			*retry = max(*retry, passwordSMSQuotaRetry(q, now))
			q.LastSend = now
			q.AttemptNonce = nonce
			q.HourCount++
			q.DayCount++
			quotas = append(quotas, q)
		}
		if *retry > 0 {
			return errs.IMErrorCode_APP_SMS_RATE_LIMITED
		}
		for _, q := range quotas {
			if e := tx.Save(&q).Error; e != nil {
				tx.AddError(e)
				return errs.IMErrorCode_APP_DEFAULT
			}
		}
		var old passwordSMSChallenge
		e := tx.Where("id=?", id).Take(&old).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			tx.AddError(e)
			return errs.IMErrorCode_APP_DEFAULT
		}
		if subtle.ConstantTimeCompare([]byte(old.Digest), []byte(s.digest(app, user, old.Phone, old.Nonce, code, old.PhoneVersion))) == 1 {
			tx.AddError(errors.New("random code collision; retry"))
			return errs.IMErrorCode_APP_DEFAULT
		}
		ch := passwordSMSChallenge{ID: id, Phone: phone, PhoneVersion: u.PasswordSMSPhoneVersion, Nonce: nonce, Digest: s.digest(app, user, phone, nonce, code, u.PasswordSMSPhoneVersion), ExpiresAt: now + 300}
		if e := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ch).Error; e != nil {
			tx.AddError(e)
			return errs.IMErrorCode_APP_DEFAULT
		}
		return errs.IMErrorCode_SUCCESS
	})
	if result != 0 {
		return result
	}
	if e := send(ctx, app, phone, code); e != nil {
		var rejected *smsengines.SMSRejectedError
		// Cleanup must survive an HTTP request cancellation after provider reply.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if errors.As(e, &rejected) {
			err := s.DB.WithContext(cleanup).Transaction(func(tx *gorm.DB) error {
				for _, key := range quotaKeys {
					if err := tx.Model(&passwordSMSQuota{}).Where("id=? AND attempt_nonce=?", key, nonce).Updates(map[string]interface{}{"last_send": 0, "attempt_nonce": ""}).Error; err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				*retry = 60
				return errs.IMErrorCode_APP_DEFAULT
			}
		}
		*retry = s.quotaRetry(cleanup, quotaKeys)
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	*retry = s.quotaRetry(context.WithoutCancel(ctx), quotaKeys)
	// A delayed provider response must not activate a replacement challenge.
	r := s.DB.WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Model(&passwordSMSChallenge{}).Where("id=? AND nonce=?", id, nonce).Update("delivered", true)
	if r.Error != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	if r.RowsAffected != 1 {
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	return errs.IMErrorCode_SUCCESS
}

func passwordSMSQuotaRetry(q passwordSMSQuota, now int64) int64 {
	var retry int64
	if q.LastSend > 0 {
		retry = max(retry, q.LastSend+60-now)
	}
	if q.HourCount >= 5 {
		retry = max(retry, q.HourStart+3600-now)
	}
	if q.DayCount >= 10 {
		retry = max(retry, q.DayStart+86400-now)
	}
	return retry
}

func (s passwordSMSService) quotaRetry(ctx context.Context, keys []string) int64 {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var quotas []passwordSMSQuota
	if err := s.DB.WithContext(ctx).Where("id IN ?", keys).Find(&quotas).Error; err != nil {
		return 60
	}
	var retry int64
	for _, q := range quotas {
		retry = max(retry, passwordSMSQuotaRetry(q, s.Now().Unix()))
	}
	return retry
}
func (s passwordSMSService) consume(ctx context.Context, code, password string) errs.IMErrorCode {
	if _, _, ok := passwordSMSIdentity(ctx); !ok {
		return errs.IMErrorCode_APP_NOT_LOGIN
	}
	if len(s.Pepper) < 32 {
		return errs.IMErrorCode_APP_PASSWORD_SMS_DISABLED
	}
	if !validInitialPassword(password) {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	app, user, _ := passwordSMSIdentity(ctx)
	return s.transaction(ctx, func(tx *gorm.DB, u *passwordSMSUser) errs.IMErrorCode {
		var ch passwordSMSChallenge
		e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", passwordSMSKey("set_password", app, user)).Take(&ch).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return errs.IMErrorCode_APP_PASSWORD_SMS_INVALID
		}
		if e != nil {
			tx.AddError(e)
			return errs.IMErrorCode_APP_DEFAULT
		}
		if ch.Phone != u.Phone || ch.PhoneVersion != u.PasswordSMSPhoneVersion || !ch.Delivered || ch.ExpiresAt <= s.Now().Unix() {
			return errs.IMErrorCode_APP_PASSWORD_SMS_INVALID
		}
		if ch.Failures >= 5 {
			return errs.IMErrorCode_APP_SMS_ATTEMPTS_EXCEEDED
		}
		digest := s.digest(app, user, u.Phone, ch.Nonce, code, u.PasswordSMSPhoneVersion)
		if code == "000000" || subtle.ConstantTimeCompare([]byte(ch.Digest), []byte(digest)) != 1 {
			ch.Failures++
			if e := tx.Model(&ch).Update("failures", ch.Failures).Error; e != nil {
				tx.AddError(e)
				return errs.IMErrorCode_APP_DEFAULT
			}
			if ch.Failures >= 5 {
				return errs.IMErrorCode_APP_SMS_ATTEMPTS_EXCEEDED
			}
			return errs.IMErrorCode_APP_PASSWORD_SMS_INVALID
		}
		r := tx.Model(&dbs.UserDao{}).Where("id=? AND phone=? AND (login_pass IS NULL OR login_pass='')", u.ID, u.Phone).Update("login_pass", utils.SHA1(password))
		if r.Error != nil {
			tx.AddError(r.Error)
			return errs.IMErrorCode_APP_DEFAULT
		}
		if r.RowsAffected != 1 {
			tx.AddError(errors.New("password state changed"))
			return errs.IMErrorCode_APP_DEFAULT
		}
		if e := tx.Delete(&ch).Error; e != nil {
			tx.AddError(e)
			return errs.IMErrorCode_APP_DEFAULT
		}
		return errs.IMErrorCode_SUCCESS
	})
}
func initialPasswordSMS() passwordSMSService {
	return passwordSMSService{DB: dbcommons.GetDb(), Now: time.Now, Prepare: prepareInitialPasswordSMS, Pepper: configures.Config.PasswordSMSPepper}
}
func SendPasswordSMS(ctx context.Context) errs.IMErrorCode { return initialPasswordSMS().send(ctx) }
func SendPasswordSMSWithRetry(ctx context.Context) (errs.IMErrorCode, int64) {
	return initialPasswordSMS().sendWithRetry(ctx)
}
func SetPasswordBySMS(ctx context.Context, code, password string) errs.IMErrorCode {
	return initialPasswordSMS().consume(ctx, code, password)
}
func sendInitialPasswordSMS(ctx context.Context, app, phone, code string) error {
	engine, err := prepareInitialPasswordSMS(ctx, app)
	if err != nil {
		return err
	}
	return engine.SmsSend(phone, map[string]interface{}{"code": code})
}

func prepareInitialPasswordSMS(ctx context.Context, app string) (smsengines.ISmsEngine, error) {
	engine, err := resolvePasswordSMSEngine(func(key string) (string, error) {
		var ext dbs.AppExtDao
		err := dbcommons.GetDb().WithContext(ctx).Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Select("app_item_value").Where("app_key=? AND app_item_key=?", app, key).Take(&ext).Error
		return ext.AppItemValue, err
	})
	if err != nil {
		return nil, err
	}
	if err := engine.(*smsengines.AliyunSmsEngine).Prepare(); err != nil {
		return nil, err
	}
	return engine, nil
}

// Only an absent purpose record permits fallback. Operators must verify that
// the fallback template is generic verification, not login/registration-only.
func resolvePasswordSMSEngine(load func(string) (string, error)) (smsengines.ISmsEngine, error) {
	value, err := load("set_password_sms_engine_conf")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		value, err = load("sms_engine_conf")
	}
	if err != nil {
		return nil, errors.New("password SMS configuration unavailable")
	}
	var conf SmsEngineConf
	if err = json.Unmarshal([]byte(value), &conf); err != nil {
		return nil, errors.New("password SMS config invalid")
	}
	// Other legacy adapters either log provider responses or do not validate the
	// provider business status. Fail closed until purpose-safe adapters exist.
	if conf.Channel != "aliyun" {
		return nil, errors.New("password SMS requires aliyun adapter")
	}
	engine := conf.Engine()
	if engine == nil {
		return nil, errors.New("password SMS config invalid")
	}
	return engine, nil
}
