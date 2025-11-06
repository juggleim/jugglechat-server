package apis

import (
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/gin-gonic/gin"
)

func ChgGroupOwner(ctx *gin.Context) {
	req := &models.GroupOwnerChgReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.ChgGroupOwner(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func AddGrpAdministrator(ctx *gin.Context) {
	req := &models.GroupAdministratorsReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.AddGroupAdministrators(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DelGrpAdministrator(ctx *gin.Context) {
	req := &models.GroupAdministratorsReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DelGroupAdministrators(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryGrpAdministrators(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	code, resp := services.QryGroupAdministrators(ctxs.ToCtx(ctx), groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func SetGroupMute(ctx *gin.Context) {
	req := &models.SetGroupMuteReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGroupMute(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SetGroupMembersMute(ctx *gin.Context) {
	req := &models.SetGroupMemberMuteReq{}
	if err := ctx.BindJSON(&req); err != nil || req.GroupId == "" || len(req.MemberIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGroupMembersMute(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SetGrpVerifyType(ctx *gin.Context) {
	req := &models.SetGroupVerifyTypeReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGroupVerifyType(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SetGrpHisMsgVisible(ctx *gin.Context) {
	req := &models.SetGroupHisMsgVisibleReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGroupHisMsgVisible(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SetGrpManagementConfs(ctx *gin.Context) {
	req := &models.GroupManagement{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGroupManagementConfs(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
