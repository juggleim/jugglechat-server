package apis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/admins/services"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
)

func AddApplication(ctx *gin.Context) {
	var req models.Application
	if err := ctx.ShouldBindJSON(&req); err != nil || req.AppKey == "" {
		ctx.JSON(http.StatusBadRequest, &errs.AdminApiErrorMsg{
			Code: errs.AdminErrorCode_ParamError,
			Msg:  "param illegal",
		})
		return
	}
	code, resp := services.AddApplication(ctxs.ToCtx(ctx), &req)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, resp)
}

func UpdApplication(ctx *gin.Context) {
	var req models.Application
	if err := ctx.ShouldBindJSON(&req); err != nil || req.AppKey == "" || req.AppId == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	code := services.UpdApplication(ctxs.ToCtx(ctx), &req)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, nil)
}

func DelApplications(ctx *gin.Context) {
	var req models.ApplicationIds
	if err := ctx.ShouldBindJSON(&req); err != nil || req.AppKey == "" || len(req.AppIds) <= 0 {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	code := services.DelApplications(ctxs.ToCtx(ctx), &req)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, nil)
}

func QryApplications(ctx *gin.Context) {
	appkey := ctx.Query("app_key")
	if appkey == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	var err error
	var page int64 = 1
	pageStr := ctx.Query("page")
	if pageStr != "" {
		page, err = tools.String2Int64(pageStr)
		if err != nil {
			page = 1
		}
	}
	var size int64 = 20
	sizeStr := ctx.Query("size")
	if sizeStr != "" {
		size, err = tools.String2Int64(sizeStr)
		if err != nil {
			size = 20
		}
	}
	isPositiveOrder := false
	orderStr := ctx.Query("order")
	if orderStr != "" {
		order, err := strconv.Atoi(orderStr)
		if err == nil && order > 0 { //0:倒序;1:正序;
			isPositiveOrder = true
		}
	}
	code, resp := services.QryApplications(ctxs.ToCtx(ctx), appkey, page, size, isPositiveOrder)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, resp)
}
