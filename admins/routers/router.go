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
	eng.Use(CorsHandler(), InjectCtx())
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

func RouteProxy(group *gin.RouterGroup) *gin.RouterGroup {
	imAdminProxy := getImAdminProxy()
	if imAdminProxy != nil {
		group.POST("/apps/active", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/create", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/list", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/info", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/configs/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/configs/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/eventsubconfig/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/eventsubconfig/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		//translate
		group.POST("/apps/translate/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/translate/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		//sms
		group.POST("/apps/sms/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/sms/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		//rtc
		group.POST("/apps/rtcconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/rtcconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/zegoconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/zegoconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/agoraconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/agoraconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/livekitconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/livekitconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})

		group.POST("/apps/iospushcer/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/iospushcer/upload", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/iospushcer/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/fcmpushconf/upload", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/fcmpushconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/androidpushconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/androidpushconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})

		group.POST("/apps/fileconf/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/fileconf/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/fileconf/switch/get", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/fileconf/switch/set", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		//logs
		group.POST("/apps/clientlogs/notify", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/clientlogs/list", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/clientlogs/download", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/serverlogs/userconnect", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/serverlogs/connect", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/serverlogs/business", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})

		//statistic
		group.GET("/apps/statistic/msg", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/useractivity", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/userreg", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/connectcount", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/maxconnectcount", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/chrmconnectcount", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/maxchrmconnectcount", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.GET("/apps/statistic/maxchrmconnectcount_v2", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})

		group.GET("/apps/sensitivewords/list", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/sensitivewords/import", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/sensitivewords/add", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		group.POST("/apps/sensitivewords/delete", func(ctx *gin.Context) {
			imAdminProxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
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

	return group
}

func CorsHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Add("Access-Control-Allow-Headers", "*")
		context.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Writer.Header().Add("Access-Control-Allow-Credentials", "true")

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
			proxy.ModifyResponse = func(resp *http.Response) error {
				resp.Header.Set("Access-Control-Allow-Origin", "*")
				resp.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
				resp.Header.Set("Access-Control-Allow-Headers", "*")
				resp.Header.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
				resp.Header.Set("Access-Control-Allow-Credentials", "true")
				return nil
			}
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "Internal error", http.StatusServiceUnavailable)
			}
			return proxy
		}
	}
	return nil
}
