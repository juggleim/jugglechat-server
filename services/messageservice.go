package services

import (
	"context"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/storages"
)

func RecallMsg(ctx context.Context, req *models.RecallMsgReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	//check right
	if !isGrpAdmin(appkey, userId, req.TargetId) {
		return errs.IMErrorCode_APP_GROUP_NORIGHT
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		code, _, err := sdk.RecallMsg(&juggleimsdk.RecallMsgReq{
			FromId:      req.FromId,
			TargetId:    req.TargetId,
			ChannelType: req.ChannelType,
			MsgId:       req.MsgId,
			MsgTime:     req.MsgTime,
			Exts:        req.Exts,
		})
		if err == nil && code == juggleimsdk.ApiCode_Success {
			return errs.IMErrorCode_SUCCESS
		}
	}

	return errs.IMErrorCode_APP_DEFAULT
}

func DelMsgs(ctx context.Context, req *models.DelHisMsgsReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	//check right
	if !isGrpAdmin(appkey, userId, req.TargetId) {
		return errs.IMErrorCode_APP_GROUP_NORIGHT
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		msgs := []*juggleimsdk.SimpleMsg{}
		for _, m := range req.Msgs {
			msgs = append(msgs, &juggleimsdk.SimpleMsg{
				MsgId:        m.MsgId,
				MsgTime:      m.MsgTime,
				MsgReadIndex: m.MsgReadIndex,
			})
		}
		code, _, err := sdk.DelMsgs(&juggleimsdk.DelMsgsReq{
			FromId:      req.FromId,
			TargetId:    req.TargetId,
			ChannelType: req.ChannelType,
			DelScope:    1,
			Msgs:        msgs,
		})
		if err == nil && code == juggleimsdk.ApiCode_Success {
			return errs.IMErrorCode_SUCCESS
		}
	}

	return errs.IMErrorCode_APP_DEFAULT
}

func isGrpAdmin(appkey, userId, groupId string) bool {
	//is creator
	grpStorage := storages.NewGroupStorage()
	grp, err := grpStorage.FindById(appkey, groupId)
	if err == nil && grp != nil && grp.CreatorId == userId {
		return true
	}
	//is admin
	grpAdminStorage := storages.NewGroupAdminStorage()
	isAdmin := grpAdminStorage.CheckAdmin(appkey, groupId, userId)
	return isAdmin
}
