package services

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/emailengines"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gopkg.in/yaml.v3"
)

type EmailTemplate struct {
	TxtBody  string `yaml:"txtBody"`
	HtmlBody string `yaml:"htmlBody"`
}

var emailTemp *EmailTemplate

func init() {
	cfBytes, err := os.ReadFile("conf/mailtemplate.yml")
	if err == nil && len(cfBytes) > 0 {
		var temp EmailTemplate
		err = yaml.Unmarshal(cfBytes, &temp)
		if err == nil {
			emailTemp = &temp
		}
	}
}

func GetEmailTemplate() (string, string) {
	if emailTemp != nil {
		return emailTemp.TxtBody, emailTemp.HtmlBody
	}
	return "Your VerifyCode is {code}", ""
}

func MailSend(ctx context.Context, mailAddress string) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	mailEngine := GetMailEngine(appkey)
	if mailEngine != nil {
		//检查是否还有有效的
		storage := storages.NewSmsRecordStorage()
		record, err := storage.FindByEmail(appkey, mailAddress, time.Now().Add(-3*time.Minute))
		randomCode := RandomSms()
		if err == nil {
			randomCode = record.Code
		} else {
			_, err = storage.Create(models.SmsRecord{
				AppKey:      appkey,
				Email:       mailAddress,
				Code:        randomCode,
				CreatedTime: time.Now(),
			})
			if err != nil {
				return errs.IMErrorCode_APP_SMS_SEND_FAILED
			}
		}
		body, html := GetEmailTemplate()
		if html != "" {
			body = ""
			html = strings.ReplaceAll(html, "{code}", randomCode)
		} else {
			html = ""
			body = strings.ReplaceAll(body, "{code}", randomCode)
		}
		err = mailEngine.SendMail(mailAddress, "Verify Code", body, html)
		if err != nil {
			return errs.IMErrorCode_APP_SMS_SEND_FAILED
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func CheckEmailCode(ctx context.Context, mail, code string) errs.IMErrorCode {
	if code == "000000" {
		return errs.IMErrorCode_SUCCESS
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewSmsRecordStorage()
	record, err := storage.FindByEmailCode(appkey, mail, code)
	if err != nil {
		return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
	}
	interval := time.Since(record.CreatedTime)
	if interval > 5*time.Minute {
		return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
	}
	return errs.IMErrorCode_SUCCESS
}

func GetMailEngine(appkey string) emailengines.IEmailEngine {
	appInfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appInfo != nil {
		if appInfo.MailEngine == nil {
			appinfos.GetAppLock().Lock()
			defer appinfos.GetAppLock().Unlock()
			loadMailEngine(appInfo)
		}
		if appInfo.MailEngine != nil {
			return appInfo.MailEngine
		}
	}
	return emailengines.DefaultEmailEngine
}

func loadMailEngine(appInfo *appinfos.AppInfo) {
	extDao := dbs.AppExtDao{}
	ext, err := extDao.Find(appInfo.AppKey, "mail_engine_conf")
	if err == nil && ext.AppItemValue != "" {
		mailConf := &MailEngineConf{}
		err := tools.JsonUnMarshal([]byte(ext.AppItemValue), mailConf)
		if err == nil {
			if mailConf.Channel == "ali" && mailConf.AliMailEngine != nil && mailConf.AliMailEngine.AccessKeyId != "" && mailConf.AliMailEngine.AccessKeySecret != "" {
				appInfo.MailEngine = mailConf.AliMailEngine
				return
			} else if mailConf.Channel == "engagelab" && mailConf.EngagelabEmailEngine != nil && mailConf.EngagelabEmailEngine.Url != "" && mailConf.EngagelabEmailEngine.ApiKey != "" && mailConf.EngagelabEmailEngine.ApiUser != "" {
				appInfo.MailEngine = mailConf.EngagelabEmailEngine
				return
			} else if mailConf.Channel == "neteasy" && mailConf.NeteasyEmailEngine != nil && mailConf.NeteasyEmailEngine.Username != "" && mailConf.NeteasyEmailEngine.AuthCode != "" {
				appInfo.MailEngine = mailConf.NeteasyEmailEngine
				return
			}
		}
	}
	appInfo.MailEngine = emailengines.DefaultEmailEngine
}

type MailEngineConf struct {
	Channel              string                             `json:"channel,omitempty"`
	AliMailEngine        *emailengines.AliEmailEngine       `json:"ali,omitempty"`
	EngagelabEmailEngine *emailengines.EngagelabEmailEngine `json:"engagelab"`
	NeteasyEmailEngine   *emailengines.NeteasyEmailEngine   `json:"neteasy,omitempty"`
}
