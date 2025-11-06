package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"
)

func GetConverConfs(ctx *gin.Context) {
	targetId := ctx.Query("target_id")
	subChannel := ctx.Query("sub_channel")
	converTypeStr := ctx.Query("conver_type")

	var converType int32 = 0
	if converTypeStr != "" {
		intVal, err := tools.String2Int64(converTypeStr)
		if err == nil && intVal > 0 {
			converType = int32(intVal)
		}
	}
	if targetId == "" || converType <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}

	code, ret := services.GetConverConfItems(ctxs.ToCtx(ctx), targetId, subChannel, converType)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, ret)
}

func SetConverConfs(ctx *gin.Context) {
	req := &models.SetConverConfsReq{}
	if err := ctx.BindJSON(req); err != nil || req.TargetId == "" || req.ConverType <= 0 || len(req.Confs) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetConverConfItem(ctxs.ToCtx(ctx), req.TargetId, req.SubChannel, int32(req.ConverType), req.Confs)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
