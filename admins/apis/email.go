package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/admins/services"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
)

func SetEmailConf(ctx *gin.Context) {
	var req models.EmailConf
	if err := ctx.ShouldBindJSON(&req); err != nil || req.AppKey == "" || req.Conf == nil {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	code := services.SetEmailConf(ctxs.ToCtx(ctx), &req)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, nil)
}

func GetEmailConf(ctx *gin.Context) {
	appkey := ctx.Query("app_key")
	if appkey == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	code, conf := services.GetEmailConf(ctxs.ToCtx(ctx), appkey)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, conf)
}
