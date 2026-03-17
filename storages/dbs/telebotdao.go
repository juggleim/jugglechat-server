package dbs

import (
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type TeleBotDao struct {
	ID          int64     `gorm:"primary_key"`
	UserId      string    `json:"user_id"`
	BotName     string    `json:"bot_name"`
	BotToken    string    `json:"bot_token"`
	Status      int       `gorm:"status"`
	CreatedTime time.Time `gorm:"created_time"`
	AppKey      string    `json:"app_key"`
}

func (bot TeleBotDao) TableName() string {
	return "telebots"
}

func (bot TeleBotDao) Create(item models.TeleBot) (int64, error) {
	add := &TeleBotDao{
		UserId:      item.UserId,
		BotName:     item.BotName,
		BotToken:    item.BotToken,
		Status:      item.Status,
		CreatedTime: time.Now(),
		AppKey:      item.AppKey,
	}
	result := dbcommons.GetDb().Create(add)
	return add.ID, result.Error
}

func (bot TeleBotDao) FindById(id int64, appkey, userId string) (*models.TeleBot, error) {
	var item TeleBotDao
	err := dbcommons.GetDb().Where("id=? and app_key=? and user_id=?", id, appkey, userId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &models.TeleBot{
		ID:          item.ID,
		UserId:      item.UserId,
		BotName:     item.BotName,
		BotToken:    item.BotToken,
		Status:      item.Status,
		CreatedTime: item.CreatedTime,
		AppKey:      item.AppKey,
	}, nil
}

func (bot TeleBotDao) BatchDel(appkey, userId string, botIds []int64) error {
	return dbcommons.GetDb().Where("app_key=? and user_id=? and id in (?)", appkey, userId, botIds).Delete(&TeleBotDao{}).Error
}

func (bot TeleBotDao) QryTeleBots(appkey, userId string, startId, limit int64) ([]*models.TeleBot, error) {
	var items []*TeleBotDao
	err := dbcommons.GetDb().Where("app_key=? and user_id=? and id>?", appkey, userId, startId).Order("id asc").Limit(int(limit)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.TeleBot{}
	for _, item := range items {
		ret = append(ret, &models.TeleBot{
			ID:          item.ID,
			UserId:      item.UserId,
			BotName:     item.BotName,
			BotToken:    item.BotToken,
			Status:      item.Status,
			CreatedTime: item.CreatedTime,
			AppKey:      item.AppKey,
		})
	}
	return ret, nil
}
