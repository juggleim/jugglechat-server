package services

import (
	"errors"

	"github.com/juggleim/jugglechat-server/commons/errs"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type smsLoginStorage interface {
	FindByPhone(string, string) (*models.User, error)
	FindByUserId(string, string) (*models.User, error)
}
type smsLegacyBinder interface {
	BindSMSLegacyPhone(string, string, string) (*models.User, error)
}

// ResolveVerifiedSMSUser must only be called after successful SMS validation.
// A nil user with SUCCESS preserves the caller's existing new-user flow.
func ResolveVerifiedSMSUser(appkey, phone string) (*models.User, errs.IMErrorCode) {
	return resolveVerifiedSMSUser(storages.NewUserStorage(), dbs.UserDao{}, appkey, phone)
}

func resolveVerifiedSMSUser(storage smsLoginStorage, binder smsLegacyBinder, appkey, phone string) (*models.User, errs.IMErrorCode) {
	user, err := storage.FindByPhone(appkey, phone)
	if err != nil {
		return nil, errs.IMErrorCode_APP_DEFAULT
	}
	// The verified number's current owner takes precedence. Legacy IDs are
	// only a fallback for numbers with no bound user.
	if user != nil {
		return user, errs.IMErrorCode_SUCCESS
	}
	user, err = storage.FindByUserId(appkey, utils.ShortMd5(phone))
	if err != nil {
		return nil, errs.IMErrorCode_APP_DEFAULT
	}
	if user == nil {
		return nil, errs.IMErrorCode_SUCCESS
	}
	user, err = binder.BindSMSLegacyPhone(appkey, user.UserId, phone)
	if errors.Is(err, models.ErrSMSPhoneConflict) {
		return nil, errs.IMErrorCode_APP_SMS_PHONE_CONFLICT
	}
	if err != nil {
		return nil, errs.IMErrorCode_APP_DEFAULT
	}
	return user, errs.IMErrorCode_SUCCESS
}
