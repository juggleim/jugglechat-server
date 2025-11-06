package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"

	"strings"
)

func TelegramBotAdd(ctx context.Context, req *apimodels.TelegramBot) (errs.IMErrorCode, *apimodels.TelegramBot) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewTeleBotStorage()
	id, err := storage.Create(models.TeleBot{
		UserId:   userId,
		BotName:  req.BotName,
		BotToken: req.BotToken,
		AppKey:   appkey,
	})
	if err == nil {
		//active telegram bot client
		ActiveTelebotProxy(ctx, &TeleBotRel{
			AppKey:   appkey,
			UserId:   userId,
			BotToken: req.BotToken,
		})
	}
	idStr, _ := utils.EncodeInt(id)
	return errs.IMErrorCode_SUCCESS, &apimodels.TelegramBot{
		BotId:   idStr,
		BotName: req.BotName,
	}
}

func TelegramBotDel(ctx context.Context, req *apimodels.TelegramBot) errs.IMErrorCode {
	return TelegramBotBatchDel(ctx, &apimodels.TelegramBotIds{
		BotIds: []string{req.BotId},
	})
}

func TelegramBotBatchDel(ctx context.Context, req *apimodels.TelegramBotIds) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	botIds := []int64{}
	for _, idStr := range req.BotIds {
		id, err := utils.DecodeInt(idStr)
		if err == nil && id > 0 {
			botIds = append(botIds, id)
		}
	}
	if len(botIds) > 0 {
		storage := storages.NewTeleBotStorage()
		for _, id := range botIds {
			bot, err := storage.FindById(id, appkey, userId)
			if err == nil && bot != nil {
				UnActiveTelebotProxy(ctx, &TeleBotRel{
					AppKey:    appkey,
					TeleBotId: "",
					BotToken:  bot.BotToken,
				})
			}
		}
		storage.BatchDel(appkey, userId, botIds)
	}
	return errs.IMErrorCode_SUCCESS
}

func QryTelegramBots(ctx context.Context, limit int64, offset string) (errs.IMErrorCode, *apimodels.TelegramBots) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewTeleBotStorage()
	var startId int64 = 0
	id, err := utils.DecodeInt(offset)
	if err == nil {
		startId = id
	}
	ret := &apimodels.TelegramBots{
		Items: []*apimodels.TelegramBot{},
	}
	bots, err := storage.QryTeleBots(appkey, userId, startId, limit)
	if err == nil {
		for _, bot := range bots {
			idStr, _ := utils.EncodeInt(bot.ID)
			ret.Offset = idStr
			ret.Items = append(ret.Items, &apimodels.TelegramBot{
				BotId:       idStr,
				BotName:     bot.BotName,
				BotToken:    bot.BotToken,
				CreatedTime: bot.CreatedTime.UnixMilli(),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func ActiveTelebotProxy(ctx context.Context, rel *TeleBotRel) {
	url := fmt.Sprintf("%s/bot-connector/telebot/add", configures.Config.BotConnector.Domain)
	headers := getBotConnectorHeaders(ctx)
	if rel.TeleBotId == "" && rel.BotToken != "" {
		arr := strings.Split(rel.BotToken, ":")
		if len(arr) >= 2 {
			rel.TeleBotId = arr[0]
		} else {
			return
		}
	}
	resp, code, err := utils.HttpDo(http.MethodPost, url, headers, utils.ToJson(rel))
	fmt.Println("activetelebot:", resp, code, err)
}

func UnActiveTelebotProxy(ctx context.Context, rel *TeleBotRel) {
	url := fmt.Sprintf("%s/bot-connector/telebot/del", configures.Config.BotConnector.Domain)
	headers := getBotConnectorHeaders(ctx)
	if rel.TeleBotId == "" && rel.BotToken != "" {
		arr := strings.Split(rel.BotToken, ":")
		if len(arr) >= 2 {
			rel.TeleBotId = arr[0]
		} else {
			return
		}
	}
	resp, code, err := utils.HttpDo(http.MethodPost, url, headers, utils.ToJson(rel))
	fmt.Println("unactivetelebot:", resp, code, err)
}

func getBotConnectorHeaders(ctx context.Context) map[string]string {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	appinfo, exist := appinfos.GetAppInfo(appkey)
	if exist {
		nonce := fmt.Sprintf("%d", rand.Int31n(10000))
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		signature := SHA1(fmt.Sprintf("%s%s%s", appinfo.AppSecret, nonce, timestamp))

		return map[string]string{
			"Content-Type": "application/json",
			"appkey":       appkey,
			"nonce":        nonce,
			"timestamp":    timestamp,
			"signature":    signature,
		}
	} else {
		return map[string]string{
			"Content-Type": "application/json",
			"appkey":       appkey,
		}
	}
}

func SHA1(s string) string {
	o := sha1.New()
	o.Write([]byte(s))
	return hex.EncodeToString(o.Sum(nil))
}

type TeleBotRel struct {
	AppKey    string `json:"app_key"`
	TeleBotId string `json:"tele_bot_id"`
	UserId    string `json:"user_id"`
	BotToken  string `json:"bot_token"`
}
