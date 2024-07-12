package dbs

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GroupDao struct {
	ID            int64     `gorm:"primary_key"`
	GroupName     string    `gorm:"group_name"`
	GroupPortrait string    `gorm:"group_portrait"`
	CreatedTime   time.Time `gorm:"created_time"`
	UpdatedTime   time.Time `gorm:"updated_time"`
	IsMute        int       `gorm:"is_mute"`
}

func (group GroupDao) TableName() string {
	return "groupinfos"
}
func (group GroupDao) Create(item GroupDao) (int64, error) {
	err := GetDb().Create(&item).Error
	return item.ID, err
}

func (group GroupDao) IsExist(groupId string) (bool, error) {
	var item GroupDao
	err := GetDb().Where("group_id=?", groupId).Take(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (group GroupDao) FindById(grpId int64) (*GroupDao, error) {
	var item GroupDao
	err := GetDb().Where("id=?", grpId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (group GroupDao) FindByIds(grpIds []int64) ([]*GroupDao, error) {
	var item []*GroupDao
	err := GetDb().Where("id in (?)", grpIds).Find(&item).Error
	return item, err
}

func (group GroupDao) Delete(grpId int64) error {
	return GetDb().Where("id=?", grpId).Delete(&GroupMemberDao{}).Error
}

func (group GroupDao) UpdateGrpInfo(id int64, groupName, groupPortrait string) error {
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
		return fmt.Errorf("do nothing")
	}
	err := GetDb().Model(&GroupDao{}).Where("id=?", id).Update(upd).Error
	return err
}
