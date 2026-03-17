package dbs

import (
	"errors"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
)

type GroupDao struct {
	ID            int64     `gorm:"primary_key"`
	GroupId       string    `gorm:"group_id"`
	GroupName     string    `gorm:"group_name"`
	GroupPortrait string    `gorm:"group_portrait"`
	CreatorId     string    `gorm:"creator_id"`
	CreatedTime   time.Time `gorm:"created_time"`
	UpdatedTime   time.Time `gorm:"updated_time"`
	AppKey        string    `gorm:"app_key"`
	IsMute        int       `gorm:"is_mute"`
}

func (group GroupDao) TableName() string {
	return "groupinfos"
}
func (group GroupDao) Create(item models.Group) error {
	err := dbcommons.GetDb().Create(&GroupDao{
		GroupId:       item.GroupId,
		GroupName:     item.GroupName,
		GroupPortrait: item.GroupPortrait,
		CreatorId:     item.CreatorId,
		CreatedTime:   time.Now(),
		UpdatedTime:   time.Now(),
		AppKey:        item.AppKey,
		IsMute:        item.IsMute,
	}).Error
	return err
}

func (group GroupDao) IsExist(appkey, groupId string) (bool, error) {
	var item GroupDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=?", appkey, groupId).Take(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (group GroupDao) FindById(appkey, groupId string) (*models.Group, error) {
	var item GroupDao
	err := dbcommons.GetDb().Where("app_key=? and group_id=?", appkey, groupId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &models.Group{
		ID:            item.ID,
		GroupId:       item.GroupId,
		GroupName:     item.GroupName,
		GroupPortrait: item.GroupPortrait,
		CreatorId:     item.CreatorId,
		CreatedTime:   item.CreatedTime,
		UpdatedTime:   item.UpdatedTime,
		AppKey:        item.AppKey,
		IsMute:        item.IsMute,
	}, nil
}

func (group GroupDao) Delete(appkey, groupId string) error {
	return dbcommons.GetDb().Where("app_key=? and group_id=?", appkey, groupId).Delete(&GroupDao{}).Error
}

func (group GroupDao) UpdateGroupMuteStatus(appkey, groupId string, isMute int32) error {
	upd := map[string]interface{}{}
	upd["is_mute"] = isMute
	return dbcommons.GetDb().Model(&GroupDao{}).Where("app_key=? and group_id=?", appkey, groupId).Updates(upd).Error
}

func (group GroupDao) UpdateGrpName(appkey, groupId, groupName, groupPortrait string) error {
	upd := map[string]interface{}{}
	if groupName != "" {
		upd["group_name"] = groupName
	}
	if groupPortrait != "" {
		upd["group_portrait"] = groupPortrait
	}
	if len(upd) > 0 {
		upd["updated_time"] = time.Now()
	} else {
		return nil
	}
	err := dbcommons.GetDb().Model(&GroupDao{}).Where("app_key=? and group_id=?", appkey, groupId).Updates(upd).Error
	return err
}

func (group GroupDao) UpdateCreatorId(appkey, groupId, creatorId string) error {
	err := dbcommons.GetDb().Model(&GroupDao{}).Where("app_key=? and group_id=?", appkey, groupId).Update("creator_id", creatorId).Error
	return err
}

func (group GroupDao) QryGroups(appkey, name string, startId, limit int64, isPositive bool) ([]*models.Group, error) {
	var items []*GroupDao
	whereStr := "app_key=?"
	params := []interface{}{appkey}
	orderBy := "id desc"
	if isPositive {
		orderBy = "id asc"
		whereStr = whereStr + " and id>?"
		params = append(params, startId)
	} else {
		if startId > 0 {
			whereStr = whereStr + " and id<?"
			params = append(params, startId)
		}
	}
	if name != "" {
		whereStr = whereStr + " and group_name like ?"
		params = append(params, "%"+name+"%")
	}
	err := dbcommons.GetDb().Where(whereStr, params...).Order(orderBy).Limit(int(limit)).Find(&items).Error
	ret := []*models.Group{}
	if err == nil {
		for _, item := range items {
			ret = append(ret, &models.Group{
				ID:            item.ID,
				GroupId:       item.GroupId,
				GroupName:     item.GroupName,
				GroupPortrait: item.GroupPortrait,
				CreatorId:     item.CreatorId,
				CreatedTime:   item.CreatedTime,
				UpdatedTime:   item.UpdatedTime,
				AppKey:        item.AppKey,
				IsMute:        item.IsMute,
			})
		}
	}
	return ret, err
}
