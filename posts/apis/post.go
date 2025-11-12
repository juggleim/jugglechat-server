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

func PostAdd(ctx *gin.Context) {
	req := apimodels.Post{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.PostAdd(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func PostUpdate(ctx *gin.Context) {
	req := apimodels.Post{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.PostUpdate(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func PostDel(ctx *gin.Context) {
	req := apimodels.PostIds{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.PostDel(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryPosts(ctx *gin.Context) {
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
	code, resp := services.QryPosts(ctxs.ToCtx(ctx), start, limit, isPositive)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func PostInfo(ctx *gin.Context) {
	postId := ctx.Query("post_id")
	if postId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.QryPostInfo(ctxs.ToCtx(ctx), postId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}
