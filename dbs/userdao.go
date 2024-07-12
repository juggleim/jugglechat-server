package dbs

import (
	"errors"
	"fmt"
)

type UserDao struct {
	ID       int64  `gorm:"primary_key"`
	Phone    string `gorm:"phone"`
	Nickname string `gorm:"nickname"`
	Avatar   string `gorm:"avatar"`
	Password string `gorm:"password"`
	Status   int    `gorm:"status"`
	ImToken  string `gorm:"im_token"`
}

func (user UserDao) TableName() string {
	return "users"
}

func (user UserDao) Create(u UserDao) (int64, error) {
	err := db.Create(&u).Error
	return u.ID, err
}
func (user UserDao) FindByUserId(userId int64) *UserDao {
	var item UserDao
	err := GetDb().Where("id=?", userId).Take(&item).Error
	if err != nil {
		return nil
	}
	return &item
}
func (user UserDao) FindByPhone(phone string) (*UserDao, error) {
	var item UserDao
	err := db.Where("phone=?", phone).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (user UserDao) CreateOrUpdate(item UserDao) error {
	return GetDb().Exec(fmt.Sprintf("INSERT INTO %s (phone,nickname,avatar,im_token)VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE nickname=?,avatar=?,im_token=?", user.TableName()), item.Phone, item.Nickname, item.Avatar, item.ImToken, item.Nickname, item.Avatar, item.ImToken).Error
}

func (user UserDao) Update(item UserDao) error {
	upds := map[string]interface{}{}
	if item.Nickname != "" {
		upds["nickname"] = item.Nickname
	}
	if item.Avatar != "" {
		upds["avatar"] = item.Avatar
	}
	if len(upds) <= 0 {
		return errors.New("no need to update")
	}
	return GetDb().Model(&UserDao{}).Where("id=?", item.ID).Update(upds).Error
}

func (user UserDao) UpdateToken(id int64, token string) error {
	return GetDb().Model(&UserDao{}).Where("id=?", id).Update("im_token", token).Error
}
