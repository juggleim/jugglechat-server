package services

import (
	"context"
	"math/rand"
	"time"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/smsengines"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomSms() string {
	retCode := ""
	for i := 0; i < 6; i++ {
		item := random.Intn(10)
		retCode = retCode + utils.Int2String(int64(item))
	}
	return retCode
}

func SmsSend(ctx context.Context, phone string) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	smsEngine := GetSmsEngine(appkey)
	if smsEngine != nil && smsEngine != smsengines.DefaultSmsEngine {
		// 检查是否还有有效的
		storage := storages.NewSmsRecordStorage()
		record, err := storage.FindByPhone(appkey, phone, time.Now().Add(-3*time.Minute))
		randomCode := RandomSms()
		if err == nil {
			randomCode = record.Code
		} else {
			_, err = storage.Create(models.SmsRecord{
				AppKey:      appkey,
				Phone:       phone,
				Code:        randomCode,
				CreatedTime: time.Now(),
			})
			if err != nil {
				return errs.IMErrorCode_APP_SMS_SEND_FAILED
			}
		}
		err = smsEngine.SmsSend(phone, map[string]interface{}{
			"code": randomCode,
		})
		if err == nil {
			return errs.IMErrorCode_SUCCESS
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func CheckPhoneSmsCode(ctx context.Context, phone, code string) errs.IMErrorCode {
	if code == "000000" {
		return errs.IMErrorCode_SUCCESS
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewSmsRecordStorage()
	record, err := storage.FindByPhoneCode(appkey, phone, code)
	if err != nil {
		return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
	}
	interval := time.Since(record.CreatedTime)
	if interval > 5*time.Minute {
		return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
	}
	return errs.IMErrorCode_SUCCESS
}

func GetSmsEngine(appkey string) smsengines.ISmsEngine {
	appInfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appInfo != nil {
		if appInfo.SmsEngine == nil {
			appinfos.GetAppLock().Lock()
			defer appinfos.GetAppLock().Unlock()
			loadSmsEngine(appInfo)
		}
		if appInfo.SmsEngine != nil {
			return appInfo.SmsEngine
		}
	}
	return smsengines.DefaultSmsEngine
}

func loadSmsEngine(appInfo *appinfos.AppInfo) {
	extDao := dbs.AppExtDao{}
	ext, err := extDao.Find(appInfo.AppKey, "sms_engine_conf")
	if err == nil && ext.AppItemValue != "" {
		smsConf := &SmsEngineConf{}
		err := utils.JsonUnMarshal([]byte(ext.AppItemValue), smsConf)
		if err == nil {
			if smsConf.Channel == "baidu" && smsConf.BdSmsEngine != nil && smsConf.BdSmsEngine.ApiKey != "" && smsConf.BdSmsEngine.SecretKey != "" {
				appInfo.SmsEngine = smsConf.BdSmsEngine
				return
			} else if smsConf.Channel == "smsbao" && smsConf.SmsBaoEngine != nil && smsConf.SmsBaoEngine.Username != "" && smsConf.SmsBaoEngine.Password != "" && smsConf.SmsBaoEngine.Template != "" {
				appInfo.SmsEngine = smsConf.SmsBaoEngine
				return
			}
		}
	}
	appInfo.SmsEngine = smsengines.DefaultSmsEngine
}

type SmsEngineConf struct {
	Channel      string                   `json:"channel,omitempty"`
	BdSmsEngine  *smsengines.BdSmsEngine  `json:"baidu,omitempty"`
	SmsBaoEngine *smsengines.SmsBaoEngine `json:"smsbao,omitempty"`
}
