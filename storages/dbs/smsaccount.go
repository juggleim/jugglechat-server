package dbs

import (
	"errors"
	"strings"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/juggleim/jugglechat-server/storages/models"
)

// CreateSMSUser classifies only the exact unique index reported by MySQL.
func (dao UserDao) CreateSMSUser(user models.User) error {
	err := dao.Create(user)
	var me *mysqlerr.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		if i := strings.LastIndex(me.Message, " for key '"); i >= 0 {
			switch strings.TrimSuffix(me.Message[i+10:], "'") {
			case "uniq_account", "users.uniq_account":
				return models.ErrAccountConflict
			case "uniq_phone", "users.uniq_phone":
				return models.ErrSMSPhoneConflict
			}
		}
	}
	return err
}
