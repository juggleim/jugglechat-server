package dbs

import (
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type GroupAdminDao struct {
	ID          int64     `gorm:"primary_key"`
	GroupId     string    `gorm:"group_id"`
	AdminId     string    `gorm:"admin_id"`
	CreatedTime time.Time `gorm:"created_time"`
	AppKey      string    `gorm:"app_key"`
}

func (admin GroupAdminDao) TableName() string {
	return "groupadmins"
}

func (admin GroupAdminDao) Upsert(item models.GroupAdmin) error {
	return dbcommons.GetDb().Exec(fmt.Sprintf("INSERT IGNORE INTO %s (app_key,group_id,admin_id)VALUES(?,?,?)", admin.TableName()), item.AppKey, item.GroupId, item.AdminId).Error
}

func (admin GroupAdminDao) QryAdmins(appkey, groupId string) ([]*models.GroupAdmin, error) {
	var items []*GroupAdminDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=?", appkey, groupId).Find(&items).Error
	ret := []*models.GroupAdmin{}
	for _, item := range items {
		ret = append(ret, &models.GroupAdmin{
			ID:          item.ID,
			GroupId:     item.GroupId,
			AdminId:     item.AdminId,
			CreatedTime: item.CreatedTime,
			AppKey:      item.AppKey,
		})
	}
	return ret, err
}

func (admin GroupAdminDao) CheckAdmin(appkey, groupId, userId string) bool {
	var item GroupAdminDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=? and admin_id=?", appkey, groupId, userId).Take(&item).Error
	if err == nil && item.AdminId == userId {
		return true
	}
	return false
}

func (admin GroupAdminDao) BatchDel(appkey, groupId string, adminIds []string) error {
	return dbcommons.GetDb().Where("app_key=? and group_id=? and admin_id in (?)", appkey, groupId, adminIds).Delete(&GroupAdminDao{}).Error
}
