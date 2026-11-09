package apis

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/accounts"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/events"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
	"github.com/juggleim/jugglechat-server/storages"
	dbModels "github.com/juggleim/jugglechat-server/storages/models"

	"image/png"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	juggleimsdk "github.com/juggleim/imserver-sdk-go"
)

func Login(ctx *gin.Context) {
	req := &models.RegisterReq{}
	if err := ctx.BindJSON(req); err != nil || req.Password == "" || (req.Account == "" && req.Phone == "" && req.Email == "") {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	if appkey == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		return
	}
	storage := storages.NewUserStorage()
	var err error
	var user *dbModels.User
	if req.Account != "" {
		user, err = storage.FindByAccount(appkey, req.Account)
	} else if req.Phone != "" {
		user, err = storage.FindByPhone(appkey, req.Phone)
	} else if req.Email != "" {
		user, err = storage.FindByEmail(appkey, req.Email)
	}
	if err != nil || user == nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_USER_NOT_EXIST)
		return
	}
	if user.LoginPass != utils.SHA1(req.Password) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_LOGIN_ERR_PASS)
		return
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk == nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		return
	}
	resp, code, _, err := sdk.Register(juggleimsdk.User{
		UserId:       user.UserId,
		Nickname:     user.Nickname,
		UserPortrait: user.UserPortrait,
	})
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
		return
	}
	if code != juggleimsdk.ApiCode(errs.IMErrorCode_SUCCESS) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode(code))
		return
	}
	authorization, tokenErr := issueLogin(ctx, user.UserId, req.DeviceParams)
	if tokenErr != nil {
		deviceError(ctx, tokenErr, nil)
		return
	}
	responses.SuccessHttpResp(ctx, &models.LoginUserResp{
		UserId:        user.UserId,
		NickName:      user.Nickname,
		Avatar:        user.UserPortrait,
		Authorization: authorization,
		ImToken:       resp.Token,
	})
}

func checkAccount(account string) bool {
	return accounts.Valid(account)
}

func Register(ctx *gin.Context) {
	req := &models.RegisterReq{}
	if err := ctx.BindJSON(req); err != nil || req.Password == "" || (req.Account == "" && req.Phone == "" && req.Email == "") {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	userId := utils.GenerateUUIDShort11()
	nickname := fmt.Sprintf("user%05d", utils.RandInt(100000))
	storage := storages.NewUserStorage()
	var err error
	if req.Account != "" {
		//check account
		if !checkAccount(req.Account) {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
			return
		}
		nickname = req.Account
		err = storage.Create(dbModels.User{
			UserId:       userId,
			Nickname:     nickname,
			LoginAccount: req.Account,
			LoginPass:    utils.SHA1(req.Password),
			AppKey:       appkey,
		})
	} else if req.Phone != "" {
		code := services.CheckPhoneSmsCode(ctxs.ToCtx(ctx), req.Phone, req.Code)
		if code != errs.IMErrorCode_SUCCESS {
			responses.ErrorHttpResp(ctx, code)
			return
		}
		err = storage.Create(dbModels.User{
			UserId:    userId,
			Nickname:  nickname,
			Phone:     req.Phone,
			LoginPass: utils.SHA1(req.Password),
			AppKey:    appkey,
		})
	} else if req.Email != "" {
		err = storage.Create(dbModels.User{
			UserId:    userId,
			Nickname:  nickname,
			Email:     req.Email,
			LoginPass: utils.SHA1(req.Password),
			AppKey:    appkey,
		})
	}
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_USER_EXISTED)
		return
	}
	events.TriggerUserRegiste(appkey, dbModels.User{
		UserId:       userId,
		Nickname:     nickname,
		LoginAccount: req.Account,
		Phone:        req.Phone,
		Email:        req.Email,
		AppKey:       appkey,
	})
	userExtStorage := storages.NewUserExtStorage()
	userExtStorage.Upsert(dbModels.UserExt{
		UserId:    userId,
		ItemKey:   models.UserExtKey_FriendVerifyType,
		ItemValue: utils.Int2String(int64(models.FriendVerifyType_NeedFriendVerify)),
		ItemType:  models.AttItemType_Setting,
		AppKey:    appkey,
	})
	responses.SuccessHttpResp(ctx, nil)
}

func SmsSend(ctx *gin.Context) {
	req := &struct {
		Phone   string `json:"phone"`
		Purpose string `json:"purpose"`
	}{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	if req.Purpose == "set_password" {
		ctx.Set("password_sms_auth", true)
		Validate(ctx)
		if ctx.IsAborted() {
			return
		}
		code, retry := services.SendPasswordSMSWithRetry(ctxs.ToCtx(ctx))
		data := gin.H{"retry_after_seconds": retry}
		if code != errs.IMErrorCode_SUCCESS {
			apiErr := errs.GetApiErrorByCode(code)
			ctx.JSON(apiErr.HttpCode, gin.H{"code": code, "msg": apiErr.Msg, "data": data})
			return
		}
		responses.SuccessHttpResp(ctx, data)
		return
	}
	if (req.Purpose != "" && req.Purpose != "login") || req.Phone == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SmsSend(ctxs.ToCtx(ctx), req.Phone)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SmsLogin(ctx *gin.Context) {
	req := &models.SmsLoginReq{}
	if err := ctx.BindJSON(req); err != nil || req.Phone == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.CheckPhoneSmsCode(ctxs.ToCtx(ctx), req.Phone, req.Code)
	if code == errs.IMErrorCode_SUCCESS {
		appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
		userId := ""
		nickname := fmt.Sprintf("user%05d", rand.Intn(100000))
		userPortrait := ""
		user, resolveCode := services.ResolveVerifiedSMSUser(appkey, req.Phone)
		if resolveCode != errs.IMErrorCode_SUCCESS {
			responses.ErrorHttpResp(ctx, resolveCode)
			return
		}
		if user == nil {
			var created bool
			user, created, resolveCode = services.CreateVerifiedSMSUser(dbModels.User{
				UserId:   utils.GenerateUUIDShort11(),
				Nickname: nickname,
				Phone:    req.Phone,
				AppKey:   appkey,
			})
			if resolveCode != errs.IMErrorCode_SUCCESS {
				responses.ErrorHttpResp(ctx, resolveCode)
				return
			}
			if created {
				events.TriggerUserRegiste(appkey, *user)
				userExtStorage := storages.NewUserExtStorage()
				userExtStorage.Upsert(dbModels.UserExt{
					UserId:    user.UserId,
					ItemKey:   models.UserExtKey_FriendVerifyType,
					ItemValue: utils.Int2String(int64(models.FriendVerifyType_NeedFriendVerify)),
					ItemType:  models.AttItemType_Setting,
					AppKey:    appkey,
				})
			}
		}
		userId = user.UserId
		nickname = user.Nickname
		userPortrait = user.UserPortrait
		sdk := imsdk.GetImSdk(appkey)
		if sdk == nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
			return
		}
		resp, code, _, err := sdk.Register(juggleimsdk.User{
			UserId:       userId,
			Nickname:     nickname,
			UserPortrait: userPortrait,
		})
		if err != nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
			return
		}
		if code != juggleimsdk.ApiCode(errs.IMErrorCode_SUCCESS) {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode(code))
			return
		}

		authorization, tokenErr := issueLogin(ctx, userId, req.DeviceParams)
		if tokenErr != nil {
			deviceError(ctx, tokenErr, nil)
			return
		}
		responses.SuccessHttpResp(ctx, &models.LoginUserResp{
			UserId:        userId,
			NickName:      nickname,
			Avatar:        userPortrait,
			Authorization: authorization,
			ImToken:       resp.Token,
		})
	} else {
		responses.ErrorHttpResp(ctx, code)
		return
	}
}

func EmailSend(ctx *gin.Context) {
	req := &models.EmailLoginReq{}
	if err := ctx.BindJSON(req); err != nil || req.Email == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.MailSend(ctxs.ToCtx(ctx), req.Email)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func EmailLogin(ctx *gin.Context) {
	req := &models.EmailLoginReq{}
	if err := ctx.BindJSON(req); err != nil || req.Email == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.CheckEmailCode(ctxs.ToCtx(ctx), req.Email, req.Code)
	if code == errs.IMErrorCode_SUCCESS {
		appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
		var userId string
		nickname := fmt.Sprintf("user%05d", utils.RandInt(100000))
		userportrait := ""
		storage := storages.NewUserStorage()
		user, err := storage.FindByEmail(appkey, req.Email)
		if err == nil && user != nil {
			userId = user.UserId
			nickname = user.Nickname
			userportrait = user.UserPortrait
		} else {
			userId = utils.GenerateUUIDShort11()
			err = storage.Create(dbModels.User{
				UserId:   userId,
				Nickname: nickname,
				Email:    req.Email,
				AppKey:   appkey,
			})
			if err != nil {
				responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
				return
			} else {
				events.TriggerUserRegiste(appkey, dbModels.User{
					UserId:   userId,
					Nickname: nickname,
					Email:    req.Email,
					AppKey:   appkey,
				})
				userExtStorage := storages.NewUserExtStorage()
				userExtStorage.Upsert(dbModels.UserExt{
					UserId:    userId,
					ItemKey:   models.UserExtKey_FriendVerifyType,
					ItemValue: utils.Int2String(int64(models.FriendVerifyType_NeedFriendVerify)),
					ItemType:  models.AttItemType_Setting,
					AppKey:    appkey,
				})
			}
		}
		sdk := imsdk.GetImSdk(appkey)
		if sdk == nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
			return
		}
		resp, code, _, err := sdk.Register(juggleimsdk.User{
			UserId:       userId,
			Nickname:     nickname,
			UserPortrait: userportrait,
		})
		if err != nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
			return
		}
		if code != juggleimsdk.ApiCode(errs.IMErrorCode_SUCCESS) {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode(code))
			return
		}

		authorization, tokenErr := issueLogin(ctx, userId, req.DeviceParams)
		if tokenErr != nil {
			deviceError(ctx, tokenErr, nil)
			return
		}
		responses.SuccessHttpResp(ctx, &models.LoginUserResp{
			UserId:        userId,
			NickName:      nickname,
			Avatar:        userportrait,
			Authorization: authorization,
			ImToken:       resp.Token,
		})
	} else {
		responses.ErrorHttpResp(ctx, code)
		return
	}
}

func GenerateQrCode(ctx *gin.Context) {
	uuidStr := utils.GenerateUUIDString()
	m := map[string]interface{}{
		"action": "login",
		"code":   uuidStr,
	}
	qrCode, _ := qr.Encode(utils.ToJson(m), qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 400, 400)
	buf := bytes.NewBuffer([]byte{})
	err := png.Encode(buf, qrCode)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_DEFAULT)
		return
	}
	secret, err := deviceStore().CreateQR(ctx.Request.Context(), ctx.GetString(string(ctxs.CtxKey_AppKey)), uuidStr, time.Now().UnixMilli())
	if err != nil {
		deviceError(ctx, err, nil)
		return
	}
	responses.SuccessHttpResp(ctx, map[string]string{
		"id":          uuidStr,
		"poll_secret": secret,
		"qr_code":     base64.StdEncoding.EncodeToString(buf.Bytes()),
	})
}

func CheckQrCode(ctx *gin.Context) {
	req := &models.QrCode{}
	if err := ctx.BindJSON(req); err != nil || req.Id == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	record, err := deviceStore().PrepareQR(ctx.Request.Context(), appkey, req.Id, req.DeviceParams, req.PollSecret, time.Now().UnixMilli())
	if err != nil {
		deviceError(ctx, err, nil)
		return
	}
	info, ok := appinfos.GetAppInfo(appkey)
	if !ok {
		deviceError(ctx, deviceauth.ErrAuth, nil)
		return
	}
	prepared := deviceauth.LoginResult{UserID: record.UserID}
	if record.Status == "confirmed" {
		userId := record.UserID
		userStorage := storages.NewUserStorage()
		user, err := userStorage.FindByUserId(appkey, userId)
		if err != nil || user == nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_USER_NOT_EXIST)
			return
		}
		sdk := imsdk.GetImSdk(appkey)
		if sdk == nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
			return
		}
		resp, code, _, err := sdk.Register(juggleimsdk.User{
			UserId:       userId,
			Nickname:     user.Nickname,
			UserPortrait: user.UserPortrait,
		})
		if err != nil {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
			return
		}
		if code != juggleimsdk.ApiCode(errs.IMErrorCode_SUCCESS) {
			responses.ErrorHttpResp(ctx, errs.IMErrorCode(code))
			return
		}
		prepared = deviceauth.LoginResult{UserID: userId, Nickname: user.Nickname, Avatar: user.UserPortrait, IMToken: resp.Token}
	}
	result, err := deviceStore().RedeemQR(ctx.Request.Context(), appkey, req.Id, req.DeviceParams, req.PollSecret, []byte(info.AppSecureKey), time.Now().UnixMilli(), prepared, func(user string) (string, error) {
		return legacyLoginToken(appkey, user, []byte(info.AppSecureKey), time.Now().UnixMilli())
	})
	if err != nil {
		deviceError(ctx, err, nil)
		return
	}
	responses.SuccessHttpResp(ctx, result)
}

func ConfirmQrCode(ctx *gin.Context) {
	req := &models.QrCode{}
	if err := ctx.BindJSON(req); err != nil || req.Id == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	userId := ctx.GetString(string(ctxs.CtxKey_RequesterId))
	err := deviceStore().ConfirmQR(ctx.Request.Context(), appkey, req.Id, userId, time.Now().UnixMilli())
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_DEFAULT)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
