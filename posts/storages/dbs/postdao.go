package dbs

import (
	"bytes"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

type PostDao struct {
	ID           int64     `gorm:"primary_key"`
	PostId       string    `gorm:"post_id"`
	Title        string    `gorm:"title"`
	Content      []byte    `gorm:"content"`
	ContentBrief string    `gorm:"content_brief"`
	PostExset    []byte    `gorm:"post_exset"`
	IsDelete     int       `gorm:"is_delete"`
	UserId       string    `gorm:"user_id"`
	CreatedTime  int64     `gorm:"created_time"`
	UpdatedTime  time.Time `gorm:"updated_time"`
	Status       int       `gorm:"status"`
	AppKey       string    `gorm:"app_key"`
}

func (post PostDao) TableName() string {
	return "posts"
}

func (post PostDao) Create(item models.Post) error {
	return dbcommons.GetDb().Create(&PostDao{
		PostId:       item.PostId,
		Title:        item.Title,
		Content:      item.Content,
		ContentBrief: item.ContentBrief,
		PostExset:    item.PostExset,
		IsDelete:     item.IsDelete,
		UserId:       item.UserId,
		CreatedTime:  item.CreatedTime,
		UpdatedTime:  time.Now(),
		Status:       item.Status,
		AppKey:       item.AppKey,
	}).Error
}

func (post PostDao) UpdatePost(appkey, postId string, title string, content []byte, contentBrief string) error {
	upd := map[string]interface{}{}
	if title != "" {
		upd["title"] = title
	}
	if len(content) <= 0 {
		upd["content"] = content
	}
	if contentBrief != "" {
		upd["content_brief"] = contentBrief
	}
	if len(upd) <= 0 {
		return nil
	}
	upd["updated_time"] = time.Now()

	err := dbcommons.GetDb().Model(&PostDao{}).Where("app_key=? and post_id=?", appkey, postId).Updates(upd).Error
	return err
}

func (post PostDao) DelPost(appkey, postId string) error {
	return dbcommons.GetDb().Where("app_key=? and post_id=?", appkey, postId).Delete(&PostDao{}).Error
}

func (post PostDao) FindById(appkey, postId string) (*models.Post, error) {
	var item PostDao
	err := dbcommons.GetDb().Where("app_key=? and post_id=?", appkey, postId).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &models.Post{
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
	}, nil
}

func (post PostDao) FindByIds(appkey string, postIds []string) (map[string]*models.Post, error) {
	var items []*PostDao
	err := dbcommons.GetDb().Where("app_key=? and post_id in (?)", appkey, postIds).Find(&items).Error
	if err != nil {
		return nil, err
	}
	ret := map[string]*models.Post{}
	for _, item := range items {
		ret[item.PostId] = &models.Post{
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
		}
	}
	return ret, nil
}

func (post PostDao) QryPosts(appkey string, startTime, limit int64, isPositive bool) ([]*models.Post, error) {
	var items []*PostDao
	conditionBuf := bytes.Buffer{}
	params := []interface{}{}
	conditionBuf.WriteString("app_key=?")
	params = append(params, appkey)
	orderStr := "created_time desc"
	if isPositive {
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
	err := dbcommons.GetDb().Where(conditionBuf.String(), params...).Order(orderStr).Limit(int(limit)).Find(&items).Error
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
