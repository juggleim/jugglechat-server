package dbs

import "time"

type SmsDao struct {
	ID          int64     `gorm:"primary_key"`
	Phone       string    `gorm:"phone"`
	Code        string    `gorm:"code"`
	CreatedTime time.Time `gorm:"created_time"`
}

func (sms SmsDao) TableName() string {
	return "sms"
}

func (sms SmsDao) Create(s SmsDao) (int64, error) {
	err := db.Create(&s).Error
	return s.ID, err
}

func (sms SmsDao) FindByPhoneCode(phone, code string) (*SmsDao, error) {
	var item SmsDao
	err := db.Where("phone=? and code=?", phone, code).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (sms SmsDao) FindByPhone(phone string, startTime time.Time) (*SmsDao, error) {
	var item SmsDao
	err := db.Where("phone=? and created_time>?", phone, startTime).Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
