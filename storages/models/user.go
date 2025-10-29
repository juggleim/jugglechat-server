package models

import "time"

type User struct {
	ID           int64
	UserId       string
	Nickname     string
	UserPortrait string
	Pinyin       string
	UserType     int
	Phone        string
	Email        string
	LoginAccount string
	LoginPass    string
	Status       int
	UpdatedTime  time.Time
	CreatedTime  time.Time
	AppKey       string
}

type IUserStorage interface {
	Create(item User) error
	Upsert(item User) error
	FindByPhone(appkey, phone string) (*User, error)
	FindByEmail(appkey, email string) (*User, error)
	FindByAccount(appkey, account string) (*User, error)
	FindByUserId(appkey, userId string) (*User, error)
	FindByUserIds(appkey string, userIds []string) (map[string]*User, error)
	SearchByKeyword(appkey string, userId, keyword string) ([]*User, error)
	Update(appkey, userId, nickname, userPortrait string) error
	UpdateAccount(appkey, userId, account string) error
	UpdatePass(appkey, userId, pass string) error
	UpdatePhone(appkey, userId, phone string) error
	UpdateEmail(appkey, userId, email string) error
	Count(appkey string) int
	CountByTime(appkey string, start, end int64) int64
	QryUsers(appkey, name string, startId, limit int64, isPositiveOrder bool) ([]*User, error)
}

type UserExt struct {
	ID          int64
	UserId      string
	ItemKey     string
	ItemValue   string
	ItemType    int
	UpdatedTime time.Time
	AppKey      string
}

type IUserExtStorage interface {
	Upsert(item UserExt) error
	BatchUpsert(items []UserExt) error
	BatchDelete(appkey, itemKey string, userIds []string) error
	QryExtFields(appkey, userId string) ([]*UserExt, error)
	QryExtFieldsByItemKeys(appkey, userId string, itemKeys []string) (map[string]*UserExt, error)
	QryExtsBaseItemKey(appkey, itemKey string, startId, limit int64) ([]*UserExt, error)
}

type BanUser struct {
	ID          int64
	UserId      string
	CreatedTime time.Time
	EndTime     int64
	ScopeKey    string
	ScopeValue  string
	Ext         string
	AppKey      string
}

type IBanUserStorage interface {
	Upsert(item BanUser) error
	FindById(appkey, userId string) ([]*BanUser, error)
	DelBanUser(appkey, userId, scopeKey string) error
	CleanBaseTime(appkey, userId string, endTime int64) error
	QryBanUsers(appkey string, limit, startId int64) ([]*BanUser, error)
}

type BlockUser struct {
	ID           int64
	UserId       string
	Nickname     string
	UserPortrait string
	UserType     int
	Pinyin       string
	BlockUserId  string
	CreatedTime  int64
	AppKey       string
}

type IBlockUserStorage interface {
	Create(item BlockUser) error
	DelBlockUser(appkey, userId, blockUserId string) error
	BatchDelBlockUsers(appkey, userId string, blockUserIds []string) error
	Find(appkey, userId, blockUserId string) (*BlockUser, error)
	FindBlockUserByIds(appkey, userId string, blockUserIds []string) ([]*BlockUser, error)
	QryBlockUsers(appkey, userId string, limit, startId int64) ([]*BlockUser, error)
}
