package dbs

import (
	"fmt"

	"github.com/juggleim/jugglechat-server/commons/dbcommons"
)

type AppNavDao struct {
	ID      int64  `gorm:"primary_key"`
	AppKey  string `gorm:"app_key"`
	AliasNo string `gorm:"alias_no"`
	AppName string `gorm:"->"` //联查apps表得到，appnavs表无此字段

	AdminUrl string `gorm:"admin_url"`
	ApiUrl   string `gorm:"api_url"`
	WsUrl    string `gorm:"ws_url"`
	AppUrl   string `gorm:"app_url"`
}

func (app AppNavDao) TableName() string {
	return "appnavs"
}

func (app AppNavDao) FindByAppkey(appkey string) (*AppNavDao, error) {
	var item AppNavDao
	err := dbcommons.GetDb().Where("app_key=?", appkey).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (app AppNavDao) QryAppNavs(startId, limit int64) ([]*AppNavDao, error) {
	var items []*AppNavDao
	err := dbcommons.GetDb().Table(fmt.Sprintf("%s as n", app.TableName())).
		Select("n.*,a.app_name as app_name").
		Joins(fmt.Sprintf("left join %s as a on a.app_key=n.app_key", AppInfoDao{}.TableName())).
		Where("n.id>?", startId).Order("n.id asc").Limit(int(limit)).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (app AppNavDao) FindByAliasNo(aliasNo string) (*AppNavDao, error) {
	var item AppNavDao
	err := dbcommons.GetDb().Where("alias_no=?", aliasNo).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
