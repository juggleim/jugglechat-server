package services

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/juggleim/jugglechat-server/commons/accounts"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type smsAccountStorage interface {
	CreateSMSUser(models.User) error
	FindByPhone(string, string) (*models.User, error)
}

// CreateVerifiedSMSUser is used only after verified SMS and a missing phone/legacy user.
// The returned user is the persisted phone owner, which may be a concurrent creator.
func CreateVerifiedSMSUser(user models.User) (*models.User, bool, errs.IMErrorCode) {
	return createVerifiedSMSUser(dbs.UserDao{}, smsAccountSuffix, user)
}

func smsAccountSuffix() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var suffix [4]byte
	for i := range suffix {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		suffix[i] = alphabet[n.Int64()]
	}
	return string(suffix[:]), nil
}

func createVerifiedSMSUser(storage smsAccountStorage, suffix func() (string, error), user models.User) (*models.User, bool, errs.IMErrorCode) {
	// Never shorten or normalize the phone to make an account fit.
	if !accounts.Valid(user.Phone) {
		return nil, false, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	user.LoginAccount = user.Phone
	user.LoginPass = ""
	for attempt := 0; attempt <= 10; attempt++ {
		err := storage.CreateSMSUser(user)
		if err == nil {
			return &user, true, errs.IMErrorCode_SUCCESS
		}
		if errors.Is(err, models.ErrSMSPhoneConflict) {
			owner, readErr := storage.FindByPhone(user.AppKey, user.Phone)
			if readErr != nil || owner == nil {
				return nil, false, errs.IMErrorCode_APP_DEFAULT
			}
			return owner, false, errs.IMErrorCode_SUCCESS
		}
		if !errors.Is(err, models.ErrAccountConflict) || attempt == 10 {
			return nil, false, errs.IMErrorCode_APP_DEFAULT
		}
		if len(user.Phone)+4 > 20 {
			return nil, false, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
		}
		value, err := suffix()
		if err != nil || len(value) != 4 || !accounts.Valid(user.Phone+value) {
			return nil, false, errs.IMErrorCode_APP_DEFAULT
		}
		user.LoginAccount = user.Phone + value
	}
	return nil, false, errs.IMErrorCode_APP_DEFAULT
}
