package apis

import (
	"encoding/base64"
	"strconv"

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

func GetServerLists(ctx *gin.Context) {
	limit := 100
	if limitStr := ctx.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			limit = 100
		} else {
			limit = l
		}
	}
	var startId int64 = 0
	if offset := ctx.Query("offset"); offset != "" {
		id, err := tools.DecodeInt(offset)
		if err == nil {
			startId = id
		}
	}
	dao := dbs.AppNavDao{}
	apps, err := dao.QryAppNavs(startId, int64(limit))
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
		return
	}
	ret := &ServerInfos{
		Items: []*ServerInfo{},
	}
	for _, app := range apps {
		idStr, _ := tools.EncodeInt(app.ID)
		ret.Offset = idStr
		imservers := []string{}
		appservers := []string{}
		if app.WsUrl != "" {
			imservers = append(imservers, app.WsUrl)
		}
		if app.AppUrl != "" {
			appservers = append(appservers, app.AppUrl)
		}
		ret.Items = append(ret.Items, &ServerInfo{
			Id:         idStr,
			AppKey:     app.AppKey,
			AppName:    app.AppName,
			AliasNo:    app.AliasNo,
			ImServers:  imservers,
			AppServers: appservers,
		})
	}
	responses.SuccessHttpResp(ctx, ret)
}

type ServerInfos struct {
	Items  []*ServerInfo `json:"items"`
	Offset string        `json:"offset"`
}

type ServerInfo struct {
	Id         string   `json:"id,omitempty"`
	AppKey     string   `json:"app_key"`
	AppName    string   `json:"app_name,omitempty"`
	AliasNo    string   `json:"alias_no,omitempty"`
	ImServers  []string `json:"im_servers"`
	AppServers []string `json:"app_servers"`
}

func tmpData(str string) string {
	bs, _ := tools.AesEncrypt([]byte(str), []byte(tools.RandStr(16)))
	return base64.URLEncoding.EncodeToString(bs)
}
