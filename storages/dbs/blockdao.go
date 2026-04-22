package dbs

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type BlockDao struct {
	ID          int64     `gorm:"primary_key"`
	UserId      string    `gorm:"user_id"`
	BlockUserId string    `gorm:"block_user_id"`
	CreatedTime time.Time `gorm:"created_time"`
	AppKey      string    `gorm:"app_key"`
}

func (block *BlockDao) TableName() string {
	return "blocks"
}

func (block BlockDao) Create(item models.BlockUser) error {
	err := dbcommons.GetDb().Create(&BlockDao{
		UserId:      item.UserId,
		BlockUserId: item.BlockUserId,
		CreatedTime: time.Now(),
		AppKey:      item.AppKey,
	}).Error
	return err
}

func (block BlockDao) DelBlockUser(appkey, userId, blockUserId string) error {
	return dbcommons.GetDb().Where("app_key=? and user_id=? and block_user_id=?", appkey, userId, blockUserId).Delete(&BlockDao{}).Error
}

func (block BlockDao) BatchDelBlockUsers(appkey, userId string, blockUserIds []string) error {
	return dbcommons.GetDb().Where("app_key=? and user_id=? and block_user_id in (?)", appkey, userId, blockUserIds).Delete(&BlockDao{}).Error
}

func (block BlockDao) Find(appkey, userId, blockUserId string) (*models.BlockUser, error) {
	var item BlockDao
	err := dbcommons.GetDb().Where("app_key=? and user_id=? and block_user_id=?", appkey, userId, blockUserId).Take(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &models.BlockUser{
		ID:          item.ID,
		UserId:      item.UserId,
		BlockUserId: item.BlockUserId,
		CreatedTime: item.CreatedTime.UnixMilli(),
		AppKey:      item.AppKey,
	}, nil
}

func (block BlockDao) FindBlockUserByIds(appkey, userId string, blockUserIds []string) ([]*models.BlockUser, error) {
	var items []*BlockDao
	err := dbcommons.GetDb().Where("app_key=? and user_id=? and block_user_id in (?)", appkey, userId, blockUserIds).Order("id asc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.BlockUser{}
	for _, item := range items {
		ret = append(ret, &models.BlockUser{
			ID:          item.ID,
			UserId:      item.UserId,
			BlockUserId: item.BlockUserId,
			CreatedTime: item.CreatedTime.UnixMilli(),
			AppKey:      item.AppKey,
		})
	}
	return ret, nil
}

type BlockUserWithUser struct {
	BlockDao
	Nickname     string `gorm:"nickname"`
	UserPortrait string `gorm:"user_portrait"`
	UserType     int    `gorm:"user_type"`
	Pinyin       string `gorm:"pinyin"`
}

func (block BlockDao) QryBlockUsers(appkey, userId string, limit, startId int64) ([]*models.BlockUser, error) {
	var items []*BlockUserWithUser
	sql := fmt.Sprintf("select b.*,u.nickname,u.user_portrait,u.user_type,u.pinyin from %s as b left join %s as u on b.app_key=u.app_key and b.block_user_id=u.user_id where b.app_key=? and b.user_id=? and b.id>? order by b.id asc limit ?", block.TableName(), UserDao{}.TableName())
	err := dbcommons.GetDb().Raw(sql, appkey, userId, startId, limit).Scan(&items).Error
	ret := []*models.BlockUser{}
	if err == nil {
		for _, item := range items {
			ret = append(ret, &models.BlockUser{
				ID:           item.ID,
				UserId:       item.UserId,
				Nickname:     item.Nickname,
				UserPortrait: item.UserPortrait,
				UserType:     item.UserType,
				Pinyin:       item.Pinyin,
				BlockUserId:  item.BlockUserId,
				CreatedTime:  item.CreatedTime.UnixMilli(),
				AppKey:       item.AppKey,
			})
		}
	}
	return ret, err
}
