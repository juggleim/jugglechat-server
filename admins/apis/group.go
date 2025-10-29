package apis

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/commons/ctxs"
	"github.com/juggleim/commons/errs"
	"github.com/juggleim/commons/responses"
	"github.com/juggleim/commons/tools"
	"github.com/juggleim/jugglechat-server/admins/services"
)

func QryGroups(ctx *gin.Context) {
	appkey := ctx.Query("app_key")
	if appkey == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	groupId := ctx.Query("group_id")
	name := ctx.Query("name")
	offset := ctx.Query("offset")
	var count int64 = 20
	var err error
	countStr := ctx.Query("count")
	if countStr != "" {
		count, err = tools.String2Int64(countStr)
		if err != nil {
			count = 20
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
	code, grps := services.QryGroups(ctxs.ToCtx(ctx), appkey, groupId, name, offset, count, isPositiveOrder)
	if code != errs.AdminErrorCode_Success {
		responses.AdminErrorHttpResp(ctx, code)
		return
	}
	responses.AdminSuccessHttpResp(ctx, grps)
}
