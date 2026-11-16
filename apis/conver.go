package apis

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/services"
)

func GetConverConfs(ctx *gin.Context) {
	targetId := ctx.Query("target_id")
	subChannel := ctx.Query("sub_channel")
	converTypeStr := ctx.Query("conver_type")

	if targetId == "" || (converTypeStr != "1" && converTypeStr != "2") {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	converType := int32(1)
	if converTypeStr == "2" {
		converType = 2
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
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	err := decoder.Decode(req)
	var extra interface{}
	if err != nil || decoder.Decode(&extra) != io.EOF || req.TargetId == "" || (req.ConverType != 1 && req.ConverType != 2) || len(req.Confs) <= 0 {
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
