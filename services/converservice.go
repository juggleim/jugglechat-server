package services

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
)

type ConverConfItemKey string

const (
	ConverConfItemKey_MsgLifeTime       ConverConfItemKey = "msg_life_time" //ms
	ConverConfItemKey_ForbidForward     ConverConfItemKey = "forbid_forward"
	ConverConfItemKey_ContentProtection ConverConfItemKey = "content_protection"
	ConverConfItemKey_WatermarkEnabled  ConverConfItemKey = "watermark_enabled"
)

func isConverBoolKey(key string) bool {
	return key == string(ConverConfItemKey_ForbidForward) || key == string(ConverConfItemKey_ContentProtection) || key == string(ConverConfItemKey_WatermarkEnabled)
}

// Preserve int64 lifetime storage without ever rounding JSON through float64.
func converLifetime(value interface{}) (int64, bool) {
	var n int64
	switch v := value.(type) {
	case json.Number:
		var err error
		n, err = v.Int64()
		if err != nil {
			return 0, false
		}
	case int:
		n = int64(v)
	case int32:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v >= 9223372036854775808.0 || math.Trunc(v) != v {
			return 0, false
		}
		n = int64(v)
	default:
		return 0, false
	}
	return n, n >= 0
}

func authorizeConverConf(ctx context.Context, target string, typ int32, write bool) errs.IMErrorCode {
	app, user := ctxs.GetAppKeyFromCtx(ctx), ctxs.GetRequesterIdFromCtx(ctx)
	kind, _ := ctx.Value(ctxs.CtxKey_AuthKind).(string)
	if strings.TrimSpace(app) == "" || strings.TrimSpace(user) == "" || (kind != "legacy" && kind != "user_session" && !(kind == "service" && !write)) {
		return errs.IMErrorCode_APP_NOT_LOGIN
	}
	if strings.TrimSpace(target) == "" || (typ != 1 && typ != 2) {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	u, err := storages.NewUserStorage().FindByUserId(app, user)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	if u == nil {
		return errs.IMErrorCode_APP_NOT_LOGIN
	}
	if typ == 1 {
		u, err = storages.NewUserStorage().FindByUserId(app, target)
		if err != nil {
			return errs.IMErrorCode_APP_DEFAULT
		}
		if u == nil {
			return errs.IMErrorCode_APP_USER_NOT_EXIST
		}
	} else {
		exists, err := storages.NewGroupStorage().IsExist(app, target)
		if err != nil {
			return errs.IMErrorCode_APP_DEFAULT
		}
		if !exists {
			return errs.IMErrorCode_APP_GROUP_NORIGHT
		}
		member, err := storages.NewGroupMemberStorage().Find(app, target, user)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.IMErrorCode_APP_GROUP_NORIGHT
		}
		if err != nil {
			return errs.IMErrorCode_APP_DEFAULT
		}
		if member == nil {
			return errs.IMErrorCode_APP_GROUP_NORIGHT
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func SetConverConfItem(ctx context.Context, targetId, subChannel string, converType int32, confItems map[string]interface{}) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	converId := tools.GetConversationId(userId, targetId, converType)
	confs := []models.ConverConf{}
	if len(confItems) == 0 {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	for key, value := range confItems {
		var valueStr string
		if isConverBoolKey(key) {
			v, ok := value.(bool)
			if !ok {
				return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
			}
			valueStr = strconv.FormatBool(v)
		} else if key == string(ConverConfItemKey_MsgLifeTime) {
			v, ok := converLifetime(value)
			if !ok {
				return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
			}
			valueStr = strconv.FormatInt(v, 10)
		} else {
			return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
		}
		confs = append(confs, models.ConverConf{
			ConverId:   converId,
			ConverType: converType,
			SubChannel: subChannel,
			ItemKey:    key,
			ItemValue:  valueStr,
			ItemType:   0,
			AppKey:     appkey,
		})
	}
	if code := authorizeConverConf(ctx, targetId, converType, true); code != errs.IMErrorCode_SUCCESS {
		return code
	}
	err := (dbs.ConverOutboxStore{DB: dbcommons.GetDb()}).Save(ctx, userId, targetId, confs)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}

func GetConverConfItems(ctx context.Context, targetId, subChannel string, converType int32) (errs.IMErrorCode, map[string]interface{}) {
	if code := authorizeConverConf(ctx, targetId, converType, false); code != errs.IMErrorCode_SUCCESS {
		return code, nil
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	converId := tools.GetConversationId(userId, targetId, converType)
	ret := map[string]interface{}{"forbid_forward": false, "content_protection": false, "watermark_enabled": false}
	confMap, meta, err := (dbs.ConverOutboxStore{DB: dbcommons.GetDb()}).ReadSnapshot(ctx, dbs.ConverScope{AppKey: appkey, ConverID: converId, ConverType: converType, SubChannel: subChannel})
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT, nil
	}
	ret["_config_meta"] = meta
	for _, conf := range confMap {
		if isConverBoolKey(conf.ItemKey) {
			ret[conf.ItemKey] = conf.ItemValue == "true"
		}
		if conf.ItemKey == string(ConverConfItemKey_MsgLifeTime) {
			var msgLifeTime int64 = 0
			i, err := tools.String2Int64(conf.ItemValue)
			if err == nil && i > 0 {
				msgLifeTime = i
			}
			ret[conf.ItemKey] = msgLifeTime
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}
