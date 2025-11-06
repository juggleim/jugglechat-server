package apis

import (
	"strings"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/gin-gonic/gin"
)

const (
	Header_RequestId     string = "request-id"
	Header_AppKey        string = "appkey"
	Header_Authorization string = "Authorization"
)

func Validate(ctx *gin.Context) {
	session := utils.GenerateUUIDShort11()
	ctx.Header(Header_RequestId, session)
	ctx.Set(string(ctxs.CtxKey_Session), session)

	//check appkey
	appkey := ctx.Request.Header.Get(Header_AppKey)
	if appkey == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_APPKEY_REQUIRED)
		ctx.Abort()
		return
	}
	ctx.Set(string(ctxs.CtxKey_AppKey), appkey)
	//check app exist
	appInfo, exist := appinfos.GetAppInfo(appkey)
	if !exist {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		ctx.Abort()
		return
	}
	urlPath := ctx.Request.URL.Path
	if urlPath != "/jim/login" && urlPath != "/jim/register" && urlPath != "/jim/sms/send" && urlPath != "/jim/sms_login" && urlPath != "/jim/sms/login" && urlPath != "/jim/email/send" && urlPath != "/jim/email/login" && urlPath != "/jim/login/qrcode" && urlPath != "/jim/login/qrcode/check" {
		//current userId
		tokenStr := ctx.Request.Header.Get(Header_Authorization)
		if tokenStr == "" {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
			ctx.Abort()
			return
		}
		if strings.HasPrefix(tokenStr, "Bearer ") {
			tokenStr = tokenStr[7:]
			if !services.CheckApiKey(tokenStr, appkey, appInfo.AppSecureKey) {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
		} else {
			authToken, err := services.ParseTokenString(tokenStr)
			if err != nil {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			token, err := services.ParseToken(authToken, []byte(appInfo.AppSecureKey))
			if err != nil {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				ctx.Abort()
				return
			}
			ctx.Set(string(ctxs.CtxKey_RequesterId), token.UserId)
		}
	}
}
