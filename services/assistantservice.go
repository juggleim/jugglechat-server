package services

import (
	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/events"
	"github.com/juggleim/jugglechat-server/storages/models"
)

var CallbackUrlPrefix string

func init() {
	events.RegisteUserRegisteEvent(func(appkey string, user models.User) {
		appinfo, exist := appinfos.GetAppInfo(appkey)
		if exist && appinfo != nil {
			if exist, obj := appinfo.GetExt("open_ai_assistant"); exist && obj != nil {
				objStr := obj.(string)
				openAssistant := tools.String2Bool(objStr)
				if openAssistant {
					//register assistant
					assistantId := "assistant_" + user.UserId
					sdk := imsdk.GetImSdk(appkey)
					sdk.RegisterBot(juggleimsdk.BotInfo{
						BotId:    assistantId,
						Nickname: user.Nickname + "'s Assistant",
						Portrait: user.UserPortrait,
						BotConf: &juggleimsdk.BotConf{
							Url: configures.Config.CallbackBaseUrl + "/" + CallbackUrlPrefix + "/assistants/msgcallback",
						},
					})
					sdk.SendPrivateMsg(juggleimsdk.Message{
						SenderId:   assistantId,
						TargetId:   user.UserId,
						MsgType:    "jg:text",
						MsgContent: `{"content":"hello"}`,
					})
				}
			}
		}
	})
}
