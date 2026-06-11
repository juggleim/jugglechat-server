package dbs

import "github.com/juggleim/jugglechat-server/commons/dbcommons"

type AppNavDao struct {
	ID      int64  `gorm:"primary_key"`
	AppKey  string `gorm:"app_key"`
	AliasNo string `gorm:"alias_no"`

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

func (app AppNavDao) FindByAliasNo(aliasNo string) (*AppNavDao, error) {
	var item AppNavDao
	err := dbcommons.GetDb().Where("alias_no=?", aliasNo).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
