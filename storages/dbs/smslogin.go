package dbs

import (
	"errors"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// BindSMSLegacyPhone only fills an unbound legacy user after SMS verification.
// The unique (app_key, phone) index arbitrates concurrent claims of the number.
func (UserDao) BindSMSLegacyPhone(appkey, userID, phone string) (*models.User, error) {
	var current UserDao
	err := dbcommons.GetDb().Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("users").Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_key=? AND user_id=?", appkey, userID).Take(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrSMSLegacyMissing
			}
			return err
		}
		if current.Phone != "" && current.Phone != phone {
			return models.ErrSMSPhoneConflict
		}
		// Even the idempotent path checks ownership; never switch to another user.
		var owner struct{ UserId string }
		err := tx.Table("users").Select("user_id").Where("app_key=? AND phone=?", appkey, phone).Take(&owner).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && owner.UserId != userID {
			return models.ErrSMSPhoneConflict
		}
		if current.Phone == phone {
			return nil
		}
		result := tx.Table("users").Where("app_key=? AND user_id=? AND (phone IS NULL OR phone='')", appkey, userID).Update("phone", phone)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return models.ErrSMSPhoneConflict
		}
		current.Phone = phone
		return nil
	})
	var me *mysqlerr.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return nil, models.ErrSMSPhoneConflict
	}
	if err != nil {
		return nil, err
	}
	return &models.User{ID: current.ID, AppKey: current.AppKey, UserId: current.UserId, Phone: current.Phone, Nickname: current.Nickname, UserPortrait: current.UserPortrait, Pinyin: current.Pinyin, UserType: current.UserType, Email: current.Email, LoginAccount: current.LoginAccount, LoginPass: current.LoginPass, Status: current.Status, CreatedTime: current.CreatedTime, UpdatedTime: current.UpdatedTime}, nil
}
