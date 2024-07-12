package dbs

import (
	"bytes"
	"fmt"
	"time"
)

type FriendDao struct {
	ID       int64 `gorm:"primary_key"`
	UserId   int64 `gorm:"user_id"`
	FriendId int64 `gorm:"friend_id"`
	Status   int   `gorm:"status"`

	CreatedTime time.Time `gorm:"created_time"`
	UpdatedTime time.Time `gorm:"updated_time"`
}

func (friend FriendDao) TableName() string {
	return "friends"
}
func (friend FriendDao) Create(item FriendDao) (int64, error) {
	err := GetDb().Create(&item).Error
	return item.ID, err
}

func (friend FriendDao) BatchCreate(items []FriendDao) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into %s (`user_id`,`friend_id`)values", friend.TableName())

	buffer.WriteString(sql)
	for i, item := range items {
		if i == len(items)-1 {
			buffer.WriteString(fmt.Sprintf("(%d,%d);", item.UserId, item.FriendId))
		} else {
			buffer.WriteString(fmt.Sprintf("(%d,%d),", item.UserId, item.FriendId))
		}
	}

	err := GetDb().Debug().Exec(buffer.String()).Error
	return err
}

func (friend FriendDao) QueryFriends(userId int64, startId, limit int64) ([]*FriendDao, error) {
	var items []*FriendDao
	err := GetDb().Where(" user_id=? and id>?", userId, startId).Order("id asc").Limit(limit).Find(&items).Error
	return items, err
}

func (friend FriendDao) CheckFriend(userId1, userId2 int64) bool {
	var item FriendDao
	err := GetDb().Where("user_id=? and friend_id=?", userId1, userId2).Take(&item).Error
	if err == nil && item.ID > 0 {
		return true
	}
	return false
}
