package apis

import (
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/gin-gonic/gin"
)

func GroupApply(ctx *gin.Context) {
	req := models.GroupInviteReq{}
	if err := ctx.BindJSON(&req); err != nil || req.GroupId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.GrpJoinApply(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func GroupInvite(ctx *gin.Context) {
	req := models.GroupInviteReq{}
	if err := ctx.BindJSON(&req); err != nil || req.GroupId == "" || len(req.MemberIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.GrpInviteMembers(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func GroupComfirm(ctx *gin.Context) {
	req := models.GroupConfirm{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.GroupConfirm(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryMyGrpApplications(ctx *gin.Context) {
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryMyGrpApplications(ctxs.ToCtx(ctx), start, int32(count), int32(order), "")
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func QryMyPendingGrpInvitations(ctx *gin.Context) {
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryMyPendingGrpInvitations(ctxs.ToCtx(ctx), start, int32(count), int32(order), "")
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func QryGrpInvitations(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryGrpInvitations(ctxs.ToCtx(ctx), start, int32(count), int32(order), groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func QryGrpPendingApplications(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryGrpPendingApplications(ctxs.ToCtx(ctx), start, int32(count), int32(order), groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func QryGrpApplications(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryGrpApplications(ctxs.ToCtx(ctx), start, int32(count), int32(order), groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}
