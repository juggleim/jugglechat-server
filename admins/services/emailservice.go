package services

import (
	"context"
	"fmt"

	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

func SetEmailConf(ctx context.Context, req *apimodels.EmailConf) errs.AdminErrorCode {
	dao := dbs.AppExtDao{}
	err := dao.Upsert(req.AppKey, "mail_engine_conf", tools.ToJson(req.Conf))
	if err != nil {
		fmt.Println("set email conf failed:", err)
	}
	return errs.AdminErrorCode_Success
}

func GetEmailConf(ctx context.Context, appkey string) (errs.AdminErrorCode, *services.MailEngineConf) {
	emailConf := &services.MailEngineConf{}
	dao := dbs.AppExtDao{}
	conf, err := dao.Find(appkey, "mail_engine_conf")
	if err == nil {
		tools.JsonUnMarshal([]byte(conf.AppItemValue), emailConf)
	} else {
		fmt.Println("get email conf failed:", err)
	}
	return errs.AdminErrorCode_Success, emailConf
}
