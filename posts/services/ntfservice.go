package services

import (
	"context"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/posts/storages/models"
)

var PostNtfSystemId string = "post_ntf"

var PostNtfMsgType string = "jgd:postnotify"

type PostNtfMsg struct {
	SponsorId   string             `json:"sponsor_id"`
	PostBusType models.PostBusType `json:"post_bus_type"`
}

func SendPostNotify(ctx context.Context, targetIds []string, notify *PostNtfMsg) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.SendSystemMsg(juggleimsdk.Message{
			SenderId:   PostNtfSystemId,
			TargetIds:  targetIds,
			MsgType:    PostNtfMsgType,
			MsgContent: tools.ToJson(notify),
		})
	}
}
