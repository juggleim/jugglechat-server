package dbs

import (
	"bytes"
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

type PostFeedDao struct {
	ID       int64  `gorm:"primary_key"`
	UserId   string `gorm:"user_id"` // User ID
	PostId   string `gorm:"post_id"`
	FeedTime int64  `gorm:"feed_time"`
	AppKey   string `gorm:"app_key"` // Application key
}

func (feed PostFeedDao) TableName() string {
	return "postfeeds"
}

func (feed PostFeedDao) Create(item models.PostFeed) error {
	return dbcommons.GetDb().Create(&PostFeedDao{
		UserId:   item.UserId,
		PostId:   item.PostId,
		FeedTime: item.FeedTime,
		AppKey:   item.AppKey,
	}).Error
}

func (feed PostFeedDao) BatchCreate(items []models.PostFeed) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("INSERT IGNORE INTO %s (app_key,user_id,post_id,feed_time)VALUES", feed.TableName())
	buffer.WriteString(sql)
	params := []interface{}{}
	for i, item := range items {
		if i == len(items)-1 {
			buffer.WriteString("(?,?,?,?)")
		} else {
			buffer.WriteString("(?,?,?,?),")
		}
		params = append(params, item.AppKey, item.UserId, item.PostId, item.FeedTime)
	}
	return dbcommons.GetDb().Exec(buffer.String(), params...).Error
}

func (feed PostFeedDao) QryPosts(appkey, userId string, startTime, limit int64, isPositive bool) ([]*models.Post, error) {
	var items []*PostDao
	params := []interface{}{}
	sql := fmt.Sprintf("select p.* from %s as f left join %s as p on p.app_key=f.app_key and p.post_id=f.post_id where f.app_key=? and f.user_id=?", feed.TableName(), (&PostDao{}).TableName())
	params = append(params, appkey)
	params = append(params, userId)
	orderStr := "f.feed_time desc"
	if isPositive {
		orderStr = "f.feed_time asc"
		sql = sql + " and f.feed_time>?"
		params = append(params, startTime)
	} else {
		if startTime <= 0 {
			startTime = time.Now().UnixMilli()
		}
		sql = sql + " and f.feed_time<?"
		params = append(params, startTime)
	}
	err := dbcommons.GetDb().Raw(sql, params...).Order(orderStr).Limit(limit).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.Post{}
	for _, item := range items {
		ret = append(ret, &models.Post{
			ID:           item.ID,
			PostId:       item.PostId,
			Title:        item.Title,
			Content:      item.Content,
			ContentBrief: item.ContentBrief,
			PostExset:    item.PostExset,
			IsDelete:     item.IsDelete,
			UserId:       item.UserId,
			CreatedTime:  item.CreatedTime,
			UpdatedTime:  item.UpdatedTime,
			Status:       item.Status,
			AppKey:       item.AppKey,
		})
	}
	return ret, nil
}
