package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/posts/apis/models"
	"github.com/juggleim/jugglechat-server/posts/services"
)

func AddPostReaction(ctx *gin.Context) {
	req := models.PostReaction{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.AddPostReaction(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DelPostReaction(ctx *gin.Context) {
	req := models.PostReaction{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DelPostReaction(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryPostReactions(ctx *gin.Context) {
	postId := ctx.Query("post_id")
	if postId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.QryPostReactions(ctxs.ToCtx(ctx), postId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}
