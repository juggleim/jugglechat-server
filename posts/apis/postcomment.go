package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	apimodels "github.com/juggleim/jugglechat-server/posts/apis/models"
	"github.com/juggleim/jugglechat-server/posts/services"
)

func PostCommentAdd(ctx *gin.Context) {
	req := apimodels.PostComment{}
	if err := ctx.BindJSON(&req); err != nil || req.PostId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.PostCommentAdd(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func PostCommentUpdate(ctx *gin.Context) {
	req := apimodels.PostComment{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.PostCommentUpdate(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func PostCommentDel(ctx *gin.Context) {
	req := apimodels.PostCommentIds{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.PostCommentDel(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryPostComments(ctx *gin.Context) {
	postId := ctx.Query("post_id")
	if postId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	var limit int64 = 20
	limitStr := ctx.Query("limit")
	var err error
	if limitStr != "" {
		limit, err = utils.String2Int64(limitStr)
		if err != nil {
			limit = 20
		}
	}
	var start int64
	startTimeStr := ctx.Query("start")
	start, err = utils.String2Int64(startTimeStr)
	if err != nil {
		start = 0
	}
	var isPositive bool = false
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err == nil {
		if order == 1 {
			isPositive = true
		}
	}
	code, resp := services.QryPostComments(ctxs.ToCtx(ctx), postId, start, limit, isPositive)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}
