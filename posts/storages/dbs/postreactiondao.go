package dbs

import (
	"bytes"
	"fmt"
	"math"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

type PostReactionDao struct {
	ID          int64  `gorm:"primary_key"`
	BusId       string `gorm:"bus_id"`
	BusType     int    `gorm:"bus_type"`
	ItemKey     string `gorm:"item_key"`
	ItemValue   string `gorm:"item_value"`
	CreatedTime int64  `gorm:"created_time"`
	UserId      string `gorm:"user_id"`
	AppKey      string `gorm:"app_key"`
}

func (reaction PostReactionDao) TableName() string {
	return "postreactions"
}

func (reaction PostReactionDao) Create(item models.PostReaction) error {
	return dbcommons.GetDb().Create(&PostReactionDao{
		BusId:       item.BusId,
		BusType:     int(item.BusType),
		ItemKey:     item.ItemKey,
		ItemValue:   item.ItemValue,
		CreatedTime: item.CreatedTime,
		UserId:      item.UserId,
		AppKey:      item.AppKey,
	}).Error
}

func (reaction PostReactionDao) Upsert(item models.PostReaction) error {
	return dbcommons.GetDb().Exec(fmt.Sprintf("INSERT INTO %s (app_key,bus_id,bus_type,item_key,item_value,user_id,created_time)VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE item_value=VALUES(item_value),created_time=VALUES(created_time)", reaction.TableName()), item.AppKey, item.BusId, item.BusType, item.ItemKey, item.ItemValue, item.UserId, item.CreatedTime).Error
}

func (reaction PostReactionDao) DelReaction(appkey, busId string, busType models.PostBusType, key string) error {
	return dbcommons.GetDb().Where("app_key=? and bus_id=? and bus_type=? and item_key=?", appkey, busId, busType, key).Delete(&PostReactionDao{}).Error
}

func (reaction PostReactionDao) QryReactions(appkey, busId string, busType models.PostBusType, startId, limit int64) ([]*models.PostReaction, error) {
	var items []*PostReactionDao
	conditionBuf := bytes.Buffer{}
	params := []interface{}{}
	conditionBuf.WriteString("app_key=? and bus_id=? and bus_type=?")
	params = append(params, appkey, busId, busType)
	orderStr := "id desc"
	if startId <= 0 {
		startId = math.MaxInt64
	}
	conditionBuf.WriteString(" and id<?")
	params = append(params, startId)
	err := dbcommons.GetDb().Where(conditionBuf.String(), params...).Order(orderStr).Limit(int(limit)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.PostReaction{}
	for _, item := range items {
		ret = append(ret, &models.PostReaction{
			BusId:       item.BusId,
			BusType:     models.PostBusType(item.BusType),
			ItemKey:     item.ItemKey,
			ItemValue:   item.ItemValue,
			CreatedTime: item.CreatedTime,
			UserId:      item.UserId,
			AppKey:      item.AppKey,
		})
	}
	return ret, nil
}
