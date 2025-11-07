package routers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/admins/apis"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
)

func RouteLogin(eng *gin.Engine, prefix string) *gin.RouterGroup {
	eng.Use(CorsHandler(prefix), InjectCtx())
	group := eng.Group("/" + prefix)
	group.Use(apis.Validate)

	group.POST("/login", apis.Login)
	group.POST("/accounts/updpass", apis.UpdPassword)
	group.POST("/accounts/add", apis.AddAccount)
	group.POST("/accounts/delete", apis.DeleteAccounts)
	group.POST("/accounts/disable", apis.DisableAccounts)
	group.POST("/accounts/bindapps", apis.BindApps)
	group.POST("/accounts/unbindapps", apis.UnBindApps)
	group.GET("/accounts/list", apis.QryAccounts)
	return group
}

var proxyPathMap map[string]string

func init() {
	proxyPathMap = map[string]string{}
	proxyPathMap["/apps/active"] = http.MethodPost
	proxyPathMap["/apps/create"] = http.MethodPost
	proxyPathMap["/apps/list"] = http.MethodGet
	proxyPathMap["/apps/info"] = http.MethodGet

	proxyPathMap["/apps/configs/set"] = http.MethodPost
	proxyPathMap["/apps/configs/get"] = http.MethodPost
	proxyPathMap["/apps/eventsubconfig/set"] = http.MethodPost
	proxyPathMap["/apps/eventsubconfig/get"] = http.MethodGet
	//translate
	proxyPathMap["/apps/translate/set"] = http.MethodPost
	proxyPathMap["/apps/translate/get"] = http.MethodGet
	//sms
	proxyPathMap["/apps/sms/set"] = http.MethodPost
	proxyPathMap["/apps/sms/get"] = http.MethodGet
	//rtc
	proxyPathMap["/apps/rtcconf/set"] = http.MethodPost
	proxyPathMap["/apps/rtcconf/get"] = http.MethodGet
	proxyPathMap["/apps/zegoconf/set"] = http.MethodPost
	proxyPathMap["/apps/zegoconf/get"] = http.MethodGet
	proxyPathMap["/apps/agoraconf/set"] = http.MethodPost
	proxyPathMap["/apps/agoraconf/get"] = http.MethodGet
	proxyPathMap["/apps/livekitconf/set"] = http.MethodPost
	proxyPathMap["/apps/livekitconf/get"] = http.MethodGet
	proxyPathMap["/apps/iospushcer/set"] = http.MethodPost
	proxyPathMap["/apps/iospushcer/upload"] = http.MethodPost
	proxyPathMap["/apps/iospushcer/get"] = http.MethodGet
	proxyPathMap["/apps/fcmpushconf/upload"] = http.MethodPost
	proxyPathMap["/apps/fcmpushconf/get"] = http.MethodGet
	proxyPathMap["/apps/androidpushconf/set"] = http.MethodPost
	proxyPathMap["/apps/androidpushconf/get"] = http.MethodGet

	proxyPathMap["/apps/fileconf/set"] = http.MethodPost
	proxyPathMap["/apps/fileconf/get"] = http.MethodGet
	proxyPathMap["/apps/fileconf/switch/get"] = http.MethodGet
	proxyPathMap["/apps/fileconf/switch/set"] = http.MethodPost
	//logs
	proxyPathMap["/apps/clientlogs/notify"] = http.MethodPost
	proxyPathMap["/apps/clientlogs/list"] = http.MethodGet
	proxyPathMap["/apps/clientlogs/download"] = http.MethodGet
	proxyPathMap["/apps/serverlogs/userconnect"] = http.MethodGet
	proxyPathMap["/apps/serverlogs/connect"] = http.MethodGet
	proxyPathMap["/apps/serverlogs/business"] = http.MethodGet

	//statistic
	proxyPathMap["/apps/statistic/msg"] = http.MethodGet
	proxyPathMap["/apps/statistic/useractivity"] = http.MethodGet
	proxyPathMap["/apps/statistic/userreg"] = http.MethodGet
	proxyPathMap["/apps/statistic/connectcount"] = http.MethodGet
	proxyPathMap["/apps/statistic/maxconnectcount"] = http.MethodGet
	proxyPathMap["/apps/statistic/chrmconnectcount"] = http.MethodGet
	proxyPathMap["/apps/statistic/maxchrmconnectcount"] = http.MethodGet
	proxyPathMap["/apps/statistic/maxchrmconnectcount_v2"] = http.MethodGet

	proxyPathMap["/apps/sensitivewords/list"] = http.MethodGet
	proxyPathMap["/apps/sensitivewords/import"] = http.MethodPost
	proxyPathMap["/apps/sensitivewords/add"] = http.MethodPost
	proxyPathMap["/apps/sensitivewords/delete"] = http.MethodPost
}

func RouteProxy(group *gin.RouterGroup) *gin.RouterGroup {
	imAdminProxy := getImAdminProxy()
	if imAdminProxy != nil {
		for path, method := range proxyPathMap {
			if method == http.MethodPost {
				group.POST(path, func(ctx *gin.Context) {
					ctx.Request.Header.Set("jchat-proxy", "1")
					imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
				})
			} else if method == http.MethodGet {
				group.GET(path, func(ctx *gin.Context) {
					ctx.Request.Header.Set("jchat-proxy", "1")
					imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
				})
			}
		}
	}
	return group
}

func Route(group *gin.RouterGroup) *gin.RouterGroup {
	group.POST("/apps/file_cred", apis.GetFileCred)
	//users
	group.GET("/apps/users/list", apis.QryUsers)
	group.POST("/apps/users/add")
	group.POST("/apps/users/update")
	group.POST("/apps/users/ban", apis.BanUsers)
	group.POST("/apps/users/unban", apis.UnBanUsers)

	//groups
	group.GET("/apps/groups/list", apis.QryGroups)
	group.POST("/apps/groups/dissolve", apis.DissolveGroup)

	//convers
	group.GET("/apps/convers/list", apis.QryConversations)
	//history msgs
	group.GET("/apps/historymsgs/list", apis.QryHistoryMsgs)

	//applications
	group.POST("/apps/applications/add", apis.AddApplication)
	group.POST("/apps/applications/update", apis.UpdApplication)
	group.POST("/apps/applications/delete", apis.DelApplications)
	group.GET("/apps/applications/list", apis.QryApplications)

	//email setting
	group.POST("/apps/email/set", apis.SetEmailConf)
	group.GET("/apps/email/get", apis.GetEmailConf)

	return group
}

func CorsHandler(prefix string) gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}

func InjectCtx() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appKey := ctx.Request.Header.Get("appkey")
		ctx.Set(string(ctxs.CtxKey_AppKey), appKey)
		ctx.Next()
	}
}

func getImAdminProxy() *httputil.ReverseProxy {
	if configures.Config.ImAdminDomain != "" {
		adminUrl, err := url.Parse(configures.Config.ImAdminDomain)
		if err == nil {
			proxy := httputil.NewSingleHostReverseProxy(adminUrl)
			proxy.Director = func(r *http.Request) {
				r.URL.Scheme = adminUrl.Scheme
				r.URL.Host = adminUrl.Host
				r.Host = adminUrl.Host

				r.Header.Set("X-Forwared-For", r.RemoteAddr)
			}
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "Internal error", http.StatusServiceUnavailable)
			}
			return proxy
		}
	}
	return nil
}
