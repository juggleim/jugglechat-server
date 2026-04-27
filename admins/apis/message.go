package apis

import (
	"github.com/gin-gonic/gin"
	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/admins/services"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
)

func QryHistoryMsgs(ctx *gin.Context) {
	appkey := ctx.Query("app_key")
	if appkey == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	channelType := ctx.Query("channel_type")
	fromId := ctx.Query("from_id")
	targetId := ctx.Query("target_id")

	startTimeStr := ctx.Query("start")
	var start int64 = 0
	if startTimeStr != "" {
		val, err := tools.String2Int64(startTimeStr)
		if err == nil && val > 0 {
			start = val
		}
	}

	countStr := ctx.Query("count")
	var count int64 = 0
	if countStr != "" {
		val, err := tools.String2Int64(countStr)
		if err == nil && val > 0 {
			count = val
		}
	}

	order := ctx.Query("order")
	var isPositive bool = false
	if order == "1" {
		isPositive = true
	}

	ret := &models.HisMsgs{
		Msgs: []*models.HisMsg{},
	}

	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		cType := juggleimsdk.ChannelType_Private
		if channelType == "1" {
			cType = juggleimsdk.ChannelType_Private
		} else if channelType == "2" {
			cType = juggleimsdk.ChannelType_Group
		}
		resp, code, _, err := sdk.QryHisMsgs(fromId, targetId, cType, start, int(count), isPositive)
		if err == nil && code == juggleimsdk.ApiCode_Success && resp != nil {
			for _, msg := range resp.Msgs {
				hisMsg := &models.HisMsg{
					Sender:     services.QryUserInfo(appkey, msg.SenderId),
					MsgId:      msg.MsgId,
					MsgTime:    msg.MsgTime,
					MsgType:    msg.MsgType,
					MsgContent: msg.MsgContent,
				}
				ret.Msgs = append(ret.Msgs, hisMsg)
			}
		}
	}
	responses.AdminSuccessHttpResp(ctx, ret)
}

func RecallHistoryMsg(ctx *gin.Context) {
	var req models.RecallHisMsgReq
	if err := ctx.BindJSON(&req); err != nil || req.AppKey == "" || req.FromId == "" || req.TargetId == "" || req.MsgId == "" || req.ChannelType == 0 {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	sdk := imsdk.GetImSdk(req.AppKey)
	if sdk != nil {
		cType := juggleimsdk.ChannelType_Private
		if req.ChannelType == 1 {
			cType = juggleimsdk.ChannelType_Private
		} else if req.ChannelType == 2 {
			cType = juggleimsdk.ChannelType_Group
		}
		code, _, err := sdk.RecallMsg(&juggleimsdk.RecallMsgReq{
			FromId:      req.FromId,
			TargetId:    req.TargetId,
			ChannelType: int32(cType),
			MsgId:       req.MsgId,
			MsgTime:     req.MsgTime,
			Exts:        req.Exts,
		})
		if err != nil {
			responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ServerErr)
			return
		}
		if code != juggleimsdk.ApiCode_Success {
			responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode(code))
			return
		}
	}
	responses.AdminSuccessHttpResp(ctx, nil)
}

func DelHistoryMsg(ctx *gin.Context) {
	var req models.DelHisMsgsReq
	if err := ctx.BindJSON(&req); err != nil || req.AppKey == "" || req.FromId == "" || req.TargetId == "" || req.ChannelType == 0 || len(req.Msgs) <= 0 {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	sdk := imsdk.GetImSdk(req.AppKey)
	if sdk != nil {
		cType := juggleimsdk.ChannelType_Private
		if req.ChannelType == 1 {
			cType = juggleimsdk.ChannelType_Private
		} else if req.ChannelType == 2 {
			cType = juggleimsdk.ChannelType_Group
		}
		code, _, err := sdk.DelMsgs(&juggleimsdk.DelMsgsReq{
			FromId:      req.FromId,
			TargetId:    req.TargetId,
			ChannelType: int32(cType),
			DelScope:    1,
			Msgs:        req.Msgs,
		})
		if err != nil {
			responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ServerErr)
			return
		}
		if code != juggleimsdk.ApiCode_Success {
			responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode(code))
			return
		}
	}
	responses.AdminSuccessHttpResp(ctx, nil)
}
