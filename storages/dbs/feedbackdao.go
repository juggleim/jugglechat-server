package dbs

import (
	"fmt"
	"time"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/models"
)

type FeedbackDao struct {
	ID          int64     `gorm:"primary_key"`
	AppKey      string    `gorm:"app_key"`
	UserId      string    `gorm:"user_id"`
	Category    string    `gorm:"category"`
	Content     []byte    `gorm:"content"`
	UpdatedTime time.Time `gorm:"updated_time"`
	CreatedTime time.Time `gorm:"created_time"`
}

func (fb FeedbackDao) TableName() string {
	return "feedbacks"
}

func (fb FeedbackDao) Create(item models.Feedback) error {
	return dbcommons.GetDb().Exec(fmt.Sprintf("INSERT INTO %s (app_key,user_id,category,content)VALUES(?,?,?,?)", fb.TableName()), item.AppKey, item.UserId, item.Category, item.Content).Error
}
