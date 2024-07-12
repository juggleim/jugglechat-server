package dbs

import (
	"bytes"
	"fmt"
	"time"
)

type GroupMemberDao struct {
	ID          int64     `gorm:"primary_key"`
	GroupId     int64     `gorm:"group_id"`
	MemberId    int64     `gorm:"member_id"`
	CreatedTime time.Time `gorm:"created_time"`
	IsMute      int       `gorm:"is_mute"`
}

func (msg GroupMemberDao) TableName() string {
	return "groupmembers"
}
func (msg GroupMemberDao) Create(item GroupMemberDao) error {
	err := GetDb().Create(&item).Error
	return err
}

func (member GroupMemberDao) BatchCreate(items []GroupMemberDao) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into %s (`group_id`,`member_id`)values", member.TableName())

	buffer.WriteString(sql)
	for i, item := range items {
		if i == len(items)-1 {
			buffer.WriteString(fmt.Sprintf("(%d,%d);", item.GroupId, item.MemberId))
		} else {
			buffer.WriteString(fmt.Sprintf("(%d,%d),", item.GroupId, item.MemberId))
		}
	}

	err := GetDb().Exec(buffer.String()).Error
	return err
}

func (member GroupMemberDao) QueryMembers(groupId int64, startId, limit int64) ([]*GroupMemberDao, error) {
	var items []*GroupMemberDao
	err := GetDb().Where(" group_id=? and id>?", groupId, startId).Order("id asc").Limit(limit).Find(&items).Error
	return items, err
}

func (member GroupMemberDao) BatchDelete(groupId int64, memberIds []int64) error {
	return GetDb().Where("group_id=? and member_id in (?)", groupId, memberIds).Delete(&GroupMemberDao{}).Error
}

func (member GroupMemberDao) DeleteByGroupId(groupId string) error {
	return GetDb().Where(" group_id=?", groupId).Delete(&GroupMemberDao{}).Error
}

func (member GroupMemberDao) UpdateMute(groupId int64, isMute int, memberIds []int64) error {
	upd := map[string]interface{}{}
	upd["is_mute"] = isMute
	return GetDb().Model(&GroupMemberDao{}).Where("group_id=? and member_id in (?)", groupId, memberIds).Update(upd).Error
}

func (member GroupMemberDao) QueryGroupsByMemberId(memberId int64, startId, limit int64) ([]*GroupMemberDao, error) {
	var items []*GroupMemberDao
	err := GetDb().Where("member_id=? and id > ?", memberId, startId).Order("id asc").Limit(limit).Find(&items).Error
	return items, err
}
