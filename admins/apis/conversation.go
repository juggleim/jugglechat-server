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

func QryConversations(ctx *gin.Context) {
	appkey := ctx.Query("app_key")
	if appkey == "" {
		responses.AdminErrorHttpResp(ctx, errs.AdminErrorCode_ParamError)
		return
	}
	startStr := ctx.Query("start")
	var start int64 = 0
	if startStr != "" {
		count, err := tools.String2Int64(startStr)
		if err == nil && count > 0 {
			start = count
		}
	}
	var count int64 = 20
	countStr := ctx.Query("count")
	if countStr != "" {
		countVal, err := tools.String2Int64(countStr)
		if err == nil && countVal > 0 {
			count = countVal
		}
	}
	var targetId *string
	targetIdStr := ctx.Query("target_id")
	if targetIdStr != "" {
		targetId = &targetIdStr
	}
	channelTypeStr := ctx.Query("channel_type")
	var channelType *int32
	if channelTypeStr != "" {
		countVal, err := tools.String2Int64(channelTypeStr)
		if err == nil && countVal > 0 {
			c := int32(countVal)
			channelType = &c
		}
	}
	ret := &models.GlobalConversations{
		Items: []*models.GlobalConversation{},
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		resp, code, _, err := sdk.QryGlobalConvers(start, int(count), targetId, channelType)
		if err == nil && code == juggleimsdk.ApiCode_Success && resp != nil {
			for _, item := range resp.Items {
				conver := &models.GlobalConversation{
					ChannelType: item.ChannelType,
					Sender:      services.QryUserInfo(appkey, item.UserId),
					Time:        item.Time,
				}
				if item.ChannelType == 1 {
					conver.Receiver = services.QryUserInfo(appkey, item.TargetId)
				} else if item.ChannelType == 2 {
					conver.Group = services.QryGroupInfo(appkey, item.TargetId)
				}
				ret.Items = append(ret.Items, conver)
			}
		}
	}
	responses.AdminSuccessHttpResp(ctx, ret)
}
