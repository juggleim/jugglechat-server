package appinfos

import (
	"sync"
	"time"

	"github.com/juggleim/jugglechat-server/commons/caches"
	"github.com/juggleim/jugglechat-server/commons/emailengines"
	"github.com/juggleim/jugglechat-server/commons/smsengines"
	"github.com/juggleim/jugglechat-server/commons/transengines"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

var appCache *caches.LruCache
var appLock *sync.RWMutex

func init() {
	appCache = caches.NewLruCacheWithAddReadTimeout("appcaches", 1000, nil, 5*time.Minute, 5*time.Minute)
	appLock = &sync.RWMutex{}
}

var PostMode_Normal int = 0
var PostMode_Friend int = 1

type AppInfo struct {
	AppName      string
	AppKey       string
	AppSecret    string
	AppSecureKey string
	AppStatus    int

	SmsEngine   smsengines.ISmsEngine
	MailEngine  emailengines.IEmailEngine
	TransEngine transengines.ITransEngine
	PostMode    int // 0:normal；1:friend

	ExtMap map[string]interface{}
}

func GetAppInfo(appkey string) (*AppInfo, bool) {
	if obj, exist := appCache.Get(appkey); exist {
		return obj.(*AppInfo), true
	} else {
		appLock.Lock()
		defer appLock.Unlock()
		if obj, exist := appCache.Get(appkey); exist {
			return obj.(*AppInfo), true
		} else {
			dao := &dbs.AppInfoDao{}
			app, err := dao.FindByAppkey(appkey)
			if err == nil && app != nil {
				info := &AppInfo{
					AppName:      app.AppName,
					AppKey:       app.AppKey,
					AppSecret:    app.AppSecret,
					AppSecureKey: app.AppSecureKey,
					AppStatus:    app.AppStatus,
					ExtMap:       make(map[string]interface{}),
				}
				appCache.Add(appkey, info)
				return info, true
			}
			return nil, false
		}
	}
}

func GetAppLock() *sync.RWMutex {
	return appLock
}

var notExist interface{} = struct{}{}

func (app *AppInfo) GetExt(key string) (bool, interface{}) {
	if val, ok := app.ExtMap[key]; ok {
		if val == notExist {
			return false, nil
		}
		return true, val
	} else {
		appLock.Lock()
		defer appLock.Unlock()
		extDao := dbs.AppExtDao{}
		exts, err := extDao.FindByItemKeys(app.AppKey, []string{key})
		if err == nil {
			for _, ext := range exts {
				if ext.AppItemKey == key {
					app.ExtMap[key] = ext.AppItemValue
					return true, ext.AppItemValue
				}
			}
		}
		app.ExtMap[key] = notExist
		return false, nil
	}
}

func (app *AppInfo) GetExtByCreator(key string, creator func(val string) interface{}) (bool, interface{}) {
	if val, ok := app.ExtMap[key]; ok {
		if val == notExist {
			return false, nil
		}
		return true, val
	} else {
		appLock.Lock()
		defer appLock.Unlock()
		extDao := dbs.AppExtDao{}
		exts, err := extDao.FindByItemKeys(app.AppKey, []string{key})
		if err == nil {
			for _, ext := range exts {
				if ext.AppItemKey == key {
					obj := creator(ext.AppItemValue)
					app.ExtMap[key] = obj
					return true, obj
				}
			}
		}
		app.ExtMap[key] = notExist
		return false, nil
	}
}
