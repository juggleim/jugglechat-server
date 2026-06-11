package apis

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

func GetServerInfo(ctx *gin.Context) {
	dao := dbs.AppNavDao{}
	appkey := ctx.Query("app_key")
	aliasNo := ctx.Query("no")
	if appkey == "" && aliasNo == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_ParamError)
		return
	}
	var app *dbs.AppNavDao
	var err error
	if appkey != "" {
		app, err = dao.FindByAppkey(appkey)
		if err != nil || app == nil {
			responses.SuccessHttpResp(ctx, map[string]string{
				"server_info": tmpData(appkey),
			})
			return
		}
	} else {
		app, err = dao.FindByAliasNo(aliasNo)
		if err != nil || app == nil {
			responses.SuccessHttpResp(ctx, map[string]string{
				"server_info": tmpData(appkey),
			})
			return
		}
	}
	imservers := []string{}
	appservers := []string{}
	if app.WsUrl != "" {
		imservers = append(imservers, app.WsUrl)
	}
	if app.AppUrl != "" {
		appservers = append(appservers, app.AppUrl)
	}
	responses.SuccessHttpResp(ctx, map[string]string{
		"server_info_plain": tools.ToJson(&ServerInfo{
			AppKey:     app.AppKey,
			ImServers:  imservers,
			AppServers: appservers,
		}),
	})
}

type ServerInfo struct {
	AppKey     string   `json:"app_key"`
	ImServers  []string `json:"im_servers"`
	AppServers []string `json:"app_servers"`
}

func tmpData(str string) string {
	bs, _ := tools.AesEncrypt([]byte(str), []byte(tools.RandStr(16)))
	return base64.URLEncoding.EncodeToString(bs)
}
