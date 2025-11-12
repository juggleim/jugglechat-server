package dbs

import (
	"bytes"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

type PostCommentDao struct {
	ID              int64     `gorm:"primary_key"`
	CommentId       string    `gorm:"comment_id"`
	PostId          string    `gorm:"post_id"`
	ParentCommentId string    `gorm:"parent_comment_id"`
	ParentUserId    string    `gorm:"parent_user_id"`
	Text            string    `gorm:"text"`
	IsDelete        int       `gorm:"is_delete"`
	UserId          string    `gorm:"user_id"`
	CreatedTime     int64     `gorm:"created_time"`
	UpdatedTime     time.Time `gorm:"updated_time"`
	Status          int       `gorm:"status"`
	AppKey          string    `gorm:"app_key"`
}

func (comment PostCommentDao) TableName() string {
	return "postcomments"
}

func (comment PostCommentDao) Create(item models.PostComment) error {
	return dbcommons.GetDb().Create(&PostCommentDao{
		CommentId:       item.CommentId,
		PostId:          item.PostId,
		ParentCommentId: item.ParentCommentId,
		ParentUserId:    item.ParentUserId,
		Text:            item.Text,
		IsDelete:        item.IsDelete,
		UserId:          item.UserId,
		CreatedTime:     item.CreatedTime,
		UpdatedTime:     time.Now(),
		Status:          item.Status,
		AppKey:          item.AppKey,
	}).Error
}

func (comment PostCommentDao) FindById(appkey, commentId string) (*models.PostComment, error) {
	var item PostCommentDao
	err := dbcommons.GetDb().Where("app_key=? and comment_id=?", appkey, commentId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &models.PostComment{
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
	}, nil
}

func (comment PostCommentDao) FindByIds(appkey string, commentIds []string) (map[string]*models.PostComment, error) {
	var items []*PostCommentDao
	err := dbcommons.GetDb().Where("app_key=? and comment_id in (?)", appkey, commentIds).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := map[string]*models.PostComment{}
	for _, item := range items {
		ret[item.CommentId] = &models.PostComment{
			ID:              item.ID,
			CommentId:       item.CommentId,
			PostId:          item.PostId,
			ParentCommentId: item.ParentCommentId,
			ParentUserId:    item.ParentUserId,
			Text:            item.Text,
			IsDelete:        item.IsDelete,
			UserId:          item.UserId,
			CreatedTime:     item.CreatedTime,
			UpdatedTime:     time.Now(),
			Status:          item.Status,
			AppKey:          item.AppKey,
		}
	}
	return ret, nil
}

func (comment PostCommentDao) UpdateComment(appkey, commentId string, text string) error {
	upd := map[string]interface{}{}
	if text != "" {
		upd["text"] = text
	}
	if len(upd) <= 0 {
		return nil
	}
	upd["updated_time"] = time.Now()
	err := dbcommons.GetDb().Model(&PostCommentDao{}).Where("app_key=? and comment_id=?", appkey, commentId).Update(upd).Error
	return err
}

func (comment PostCommentDao) DelComment(appkey, commentId string) error {
	return dbcommons.GetDb().Where("app_key=? and comment_id=?", appkey, commentId).Delete(&PostCommentDao{}).Error
}

func (comment PostCommentDao) QryPostComments(appkey, postId string, startTime, limit int64, isPosttive bool) ([]*models.PostComment, error) {
	var items []*PostCommentDao
	conditionBuf := bytes.Buffer{}
	params := []interface{}{}
	conditionBuf.WriteString("app_key=? and post_id=?")
	params = append(params, appkey)
	params = append(params, postId)
	orderStr := "created_time desc"
	if isPosttive {
		orderStr = "created_time asc"
		conditionBuf.WriteString(" and created_time>?")
		params = append(params, startTime)
	} else {
		if startTime <= 0 {
			startTime = time.Now().UnixMilli()
		}
		conditionBuf.WriteString(" and created_time<?")
		params = append(params, startTime)
	}
	err := dbcommons.GetDb().Where(conditionBuf.String(), params...).Order(orderStr).Limit(limit).Find(&items).Error
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
