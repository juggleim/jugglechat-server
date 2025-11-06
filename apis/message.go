package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/services"
)

func RecallMsg(ctx *gin.Context) {
	req := &models.RecallMsgReq{}
	if err := ctx.BindJSON(req); err != nil || req.TargetId == "" || req.ChannelType != 2 || req.MsgId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.RecallMsg(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DelMsgs(ctx *gin.Context) {
	req := &models.DelHisMsgsReq{}
	if err := ctx.BindJSON(req); err != nil || req.TargetId == "" || req.ChannelType != 2 || len(req.Msgs) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DelMsgs(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
