package apis

import (
	"strings"
	"time"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"

	"github.com/gin-gonic/gin"
)

const (
	Header_RequestId     string = "request-id"
	Header_AppKey        string = "appkey"
	Header_Authorization string = "Authorization"
	Header_Version       string = "version"
)

func Validate(ctx *gin.Context) {
	// Identity kind is established only after credential verification.
	ctx.Set(string(ctxs.CtxKey_AuthKind), "")
	ctx.Set(string(ctxs.CtxKey_RequesterId), "")
	session := utils.GenerateUUIDShort11()
	ctx.Header(Header_RequestId, session)
	ctx.Set(string(ctxs.CtxKey_Session), session)
	//&& urlPath != "/jim/serverinfos"
	urlPath := ctx.Request.URL.Path
	if urlPath == "/jim/serverinfos" || urlPath == "/jim/servers/info" || urlPath == "/jim/servers/list" {
		return
	}
	//check appkey
	appkey := ctx.Request.Header.Get(Header_AppKey)
	if appkey == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_APPKEY_REQUIRED)
		ctx.Abort()
		return
	}
	ctx.Set(string(ctxs.CtxKey_AppKey), appkey)
	version := ctx.Request.Header.Get(Header_Version)
	if version != "" {
		ctx.Set(string(ctxs.CtxKey_Version), version)
	}
	//check app exist
	appInfo, exist := appinfos.GetAppInfo(appkey)
	if !exist {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		ctx.Abort()
		return
	}

	if urlPath != "/jim/login" && urlPath != "/jim/register" && urlPath != "/jim/sms/send" && urlPath != "/jim/sms_login" && urlPath != "/jim/sms/login" && urlPath != "/jim/email/send" && urlPath != "/jim/email/login" && urlPath != "/jim/login/qrcode" && urlPath != "/jim/login/qrcode/check" {
		//current userId
		tokenStr := ctx.Request.Header.Get(Header_Authorization)
		if tokenStr == "" {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
			ctx.Abort()
			return
		}
		if strings.HasPrefix(tokenStr, "Bearer ") {
			if urlPath == "/jim/users/setaccount" {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			if !deviceauth.Allowed("api_key", urlPath, deviceauth.Required(appkey, time.Now().UnixMilli())) {
				deviceError(ctx, deviceauth.ErrAuth, nil)
				ctx.Abort()
				return
			}
			tokenStr = tokenStr[7:]
			if !services.CheckApiKey(tokenStr, appkey, appInfo.AppSecureKey) {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
		} else if strings.HasPrefix(tokenStr, "svc") {
			if urlPath == "/jim/users/setaccount" {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			claims, err := deviceauth.VerifyService(tokenStr, []byte(appInfo.AppSecureKey), appkey, urlPath, time.Now().UnixMilli())
			if err != nil {
				deviceError(ctx, deviceauth.ErrAuth, nil)
				ctx.Abort()
				return
			}
			ctx.Set(string(ctxs.CtxKey_RequesterId), claims.UserID)
		} else if strings.HasPrefix(tokenStr, "ds") {
			claims, err := deviceauth.Verify(tokenStr, []byte(appInfo.AppSecureKey), appkey)
			if err == nil {
				err = deviceStore().Authenticate(ctx.Request.Context(), claims, time.Now().UnixMilli())
			}
			if err != nil {
				deviceError(ctx, deviceauth.ErrAuth, nil)
				ctx.Abort()
				return
			}
			ctx.Set("device_session", claims)
			ctx.Set(string(ctxs.CtxKey_AuthKind), "user_session")
			ctx.Set(string(ctxs.CtxKey_RequesterId), claims.UserID)
		} else {
			authToken, err := services.ParseTokenString(tokenStr)
			if err != nil || authToken.Appkey != appkey {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			token, err := services.ParseToken(authToken, []byte(appInfo.AppSecureKey))
			if err != nil || token.DeviceId != "" || !deviceauth.Allowed("legacy", urlPath, deviceauth.Required(appkey, time.Now().UnixMilli())) {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			ctx.Set(string(ctxs.CtxKey_RequesterId), token.UserId)
			ctx.Set(string(ctxs.CtxKey_AuthKind), "legacy")
			deviceauth.LegacyAuthentications.Add(1)
		}
	}
}
