package services

import (
	"context"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	utils "github.com/juggleim/jugglechat-server/commons/tools"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
)

func SendGrpNotify(ctx context.Context, grpId string, notify *apimodels.GroupNotify) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendGroupMsg(juggleimsdk.Message{
			SenderId:   requestId,
			TargetIds:  []string{grpId},
			MsgType:    apimodels.GroupNotifyMsgType,
			MsgContent: utils.ToJson(notify),
			IsStorage:  utils.BoolPtr(true),
			IsCount:    utils.BoolPtr(false),
		})
	}
}

func SendFriendNotify(ctx context.Context, targetId string, notify *apimodels.FriendNotify) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendPrivateMsg(juggleimsdk.Message{
			SenderId:   requestId,
			TargetIds:  []string{targetId},
			MsgType:    apimodels.FriendNotifyMsgType,
			MsgContent: utils.ToJson(notify),
			IsStorage:  utils.BoolPtr(true),
			IsCount:    utils.BoolPtr(true),
		})
	}
}

func SendFriendApplyNotify(ctx context.Context, targetId string, notify *apimodels.FriendApplyNotify) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendSystemMsg(juggleimsdk.Message{
			SenderId:   apimodels.SystemFriendApplyConverId,
			TargetIds:  []string{targetId},
			MsgType:    apimodels.FriendApplicationMsgType,
			MsgContent: utils.ToJson(notify),
		})
	}
}

func SendPriMsg(ctx context.Context, senderId, targetId string, msgType string, msg interface{}) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendPrivateMsg(juggleimsdk.Message{
			SenderId:   requestId,
			TargetIds:  []string{targetId},
			MsgType:    msgType,
			MsgContent: utils.ToJson(msg),
		})
	}
}

func SendGroupMsg(ctx context.Context, senderId, targetId string, msgType string, msg interface{}, mentionInfo *juggleimsdk.MentionInfo) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendGroupMsg(juggleimsdk.Message{
			SenderId:    requestId,
			TargetIds:   []string{targetId},
			MsgType:     msgType,
			MsgContent:  utils.ToJson(msg),
			MentionInfo: mentionInfo,
		})
	}
}
