package dbs

import (
	"bytes"
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

type PostCommentFeedDao struct {
	ID        int64  `gorm:"primary_key"`
	UserId    string `gorm:"user_id"` // User ID
	CommentId string `gorm:"comment_id"`
	PostId    string `gorm:"post_id"`
	FeedTime  int64  `gorm:"feed_time"`
	AppKey    string `gorm:"app_key"` // Application key
}

func (feed PostCommentFeedDao) TableName() string {
	return "postcommentfeeds"
}

func (feed PostCommentFeedDao) Create(item models.PostCommentFeed) error {
	return dbcommons.GetDb().Create(&PostCommentFeedDao{
		UserId:    item.UserId,
		CommentId: item.CommentId,
		PostId:    item.PostId,
		FeedTime:  item.FeedTime,
		AppKey:    item.AppKey,
	}).Error
}

func (feed PostCommentFeedDao) BatchCreate(items []models.PostCommentFeed) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("INSERT IGNORE INTO %s (app_key,user_id,comment_id,post_id,feed_time)VALUES", feed.TableName())
	buffer.WriteString(sql)
	params := []interface{}{}
	for i, item := range items {
		if i == len(items)-1 {
			buffer.WriteString("(?,?,?,?,?)")
		} else {
			buffer.WriteString("(?,?,?,?,?),")
		}
		params = append(params, item.AppKey, item.UserId, item.CommentId, item.PostId, item.FeedTime)
	}
	return dbcommons.GetDb().Exec(buffer.String(), params...).Error
}

func (feed PostCommentFeedDao) QryPostComments(appkey, userId, postId string, startTime, limit int64, isPositive bool) ([]*models.PostComment, error) {
	var items []*PostCommentDao
	params := []interface{}{}
	sql := fmt.Sprintf("select c.* from %s as c left join %s as f on c.app_key=f.app_key and c.comment_id=f.comment_id where f.app_key=? and f.user_id=? and f.post_id=?", (&PostCommentDao{}).TableName(), feed.TableName())
	params = append(params, appkey)
	params = append(params, userId)
	params = append(params, postId)
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
	err := dbcommons.GetDb().Raw(sql, params...).Order(orderStr).Limit(int(limit)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := []*models.PostComment{}
	for _, item := range items {
		ret = append(ret, &models.PostComment{
			ID:              item.ID,
			CommentId:       item.CommentId,
			PostId:          item.PostId,
			ParentCommentId: item.ParentCommentId,
			ParentUserId:    item.ParentUserId,
			Text:            item.Text,
			IsDelete:        item.IsDelete,
			UserId:          item.UserId,
			CreatedTime:     item.CreatedTime,
			UpdatedTime:     item.UpdatedTime,
			Status:          item.Status,
			AppKey:          item.AppKey,
		})
	}
	return ret, nil
}
