package dbs

import (
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type UserBanScope string

const (
	UserBanScopeDefault  UserBanScope = "default"
	UserBanScopePlatform UserBanScope = "platform"
	UserBanScopeDevice   UserBanScope = "device"
	UserBanScopeIp       UserBanScope = "ip"
)

type BanUserDao struct {
	ID          int64     `gorm:"primary_key"`
	UserId      string    `gorm:"user_id"`
	CreatedTime time.Time `gorm:"created_time"`
	EndTime     int64     `gorm:"end_time"`
	ScopeKey    string    `gorm:"scope_key"`
	ScopeValue  string    `gorm:"scope_value"`
	Ext         string    `gorm:"ext"`
	AppKey      string    `gorm:"app_key"`
}

func (user BanUserDao) TableName() string {
	return "banusers"
}

func (user BanUserDao) Upsert(item models.BanUser) error {
	err := dbcommons.GetDb().Exec("INSERT INTO banusers (user_id, end_time, scope_key, scope_value, ext, app_key)VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE end_time=?, scope_value=?, ext=?",
		item.UserId, item.EndTime, item.ScopeKey, item.ScopeValue, item.Ext, item.AppKey, item.EndTime, item.ScopeValue, item.Ext).Error
	return err
}

func (user BanUserDao) FindById(appkey, userId string) ([]*models.BanUser, error) {
	var items []*BanUserDao
	err := dbcommons.GetDb().Where("app_key=? and user_id=?", appkey, userId).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.BanUser{}
	for _, item := range items {
		ret = append(ret, &models.BanUser{
			ID:          item.ID,
			UserId:      item.UserId,
			CreatedTime: item.CreatedTime,
			EndTime:     item.EndTime,
			ScopeKey:    item.ScopeKey,
			ScopeValue:  item.ScopeValue,
			Ext:         item.Ext,
			AppKey:      item.AppKey,
		})
	}
	return ret, nil
}

func (user BanUserDao) DelBanUser(appkey, userId, scopeKey string) error {
	if scopeKey == "" {
		return dbcommons.GetDb().Where("app_key=? and user_id=?", appkey, userId).Delete(&BanUserDao{}).Error
	}
	return dbcommons.GetDb().Where("app_key=? and user_id=? and scope_key=?", appkey, userId, scopeKey).Delete(&BanUserDao{}).Error
}

func (user BanUserDao) CleanBaseTime(appkey, userId string, endTime int64) error {
	return dbcommons.GetDb().Where("app_key=? and user_id=? and end_time>0 and end_time<?", appkey, user, endTime).Delete(&BanUserDao{}).Error
}

func (user BanUserDao) QryBanUsers(appkey string, limit, startId int64) ([]*models.BanUser, error) {
	var items []*BanUserDao
	err := dbcommons.GetDb().Where("app_key=? and (end_time=0 or end_time>?) and id>?", appkey, time.Now().UnixMilli(), startId).Order("id asc").Limit(int(limit)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.BanUser{}
	for _, item := range items {
		ret = append(ret, &models.BanUser{
			ID:          item.ID,
			UserId:      item.UserId,
			CreatedTime: item.CreatedTime,
			EndTime:     item.EndTime,
			ScopeKey:    item.ScopeKey,
			ScopeValue:  item.ScopeValue,
			Ext:         item.Ext,
			AppKey:      item.AppKey,
		})
	}
	return ret, nil
}
