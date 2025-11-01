package apis

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/juggleim/commons/ctxs"
	"github.com/juggleim/commons/errs"
	"github.com/juggleim/commons/imsdk"
	"github.com/juggleim/commons/responses"
	utils "github.com/juggleim/commons/tools"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/events"
	"github.com/juggleim/jugglechat-server/services"
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
	responses.SuccessHttpResp(ctx, &models.LoginUserResp{
		UserId:        user.UserId,
		NickName:      user.Nickname,
		Avatar:        user.UserPortrait,
		Authorization: services.GenerateToken(appkey, user.UserId),
		ImToken:       resp.Token,
	})
}

var accountRegex *regexp.Regexp

func init() {
	accountRegex = regexp.MustCompile(`^[a-zA-Z0-9]{6,20}$`)
}

func checkAccount(account string) bool {
	return accountRegex.MatchString(account)
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
	events.TriggerUserRegiste(dbModels.User{
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
	req := &models.SmsLoginReq{}
	if err := ctx.BindJSON(req); err != nil || req.Phone == "" {
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
		userId := utils.ShortMd5(req.Phone)
		nickname := fmt.Sprintf("user%05d", utils.RandInt(100000))
		userPortrait := ""
		storage := storages.NewUserStorage()
		user, err := storage.FindByPhone(appkey, req.Phone)
		if err == nil && user != nil {
			userId = user.UserId
			nickname = user.Nickname
			userPortrait = user.UserPortrait
		} else {
			user, err = storage.FindByUserId(appkey, userId)
			if err == nil && user != nil {
				userId = user.UserId
				nickname = user.Nickname
				userPortrait = user.UserPortrait
			} else {
				userId = utils.GenerateUUIDShort11()
				err = storage.Create(dbModels.User{
					UserId:   userId,
					Nickname: nickname,
					Phone:    req.Phone,
					AppKey:   appkey,
				})
				if err != nil {
					responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_LOGIN)
					return
				} else {
					events.TriggerUserRegiste(dbModels.User{
						UserId:   userId,
						Nickname: nickname,
						Phone:    req.Phone,
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
		}
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

		responses.SuccessHttpResp(ctx, &models.LoginUserResp{
			UserId:        userId,
			NickName:      nickname,
			Avatar:        userPortrait,
			Authorization: services.GenerateToken(appkey, userId),
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
				events.TriggerUserRegiste(dbModels.User{
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

		responses.SuccessHttpResp(ctx, &models.LoginUserResp{
			UserId:        userId,
			NickName:      nickname,
			Avatar:        userportrait,
			Authorization: services.GenerateToken(appkey, userId),
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
	storage := storages.NewQrCodeRecordStorage()
	storage.Create(dbModels.QrCodeRecord{
		CodeId:      uuidStr,
		AppKey:      ctx.GetString(string(ctxs.CtxKey_AppKey)),
		CreatedTime: time.Now().UnixMilli(),
	})
	responses.SuccessHttpResp(ctx, map[string]string{
		"id":      uuidStr,
		"qr_code": base64.StdEncoding.EncodeToString(buf.Bytes()),
	})
}

func CheckQrCode(ctx *gin.Context) {
	req := &models.QrCode{}
	if err := ctx.BindJSON(req); err != nil || req.Id == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	storage := storages.NewQrCodeRecordStorage()
	record, err := storage.FindById(ctx.GetString(string(ctxs.CtxKey_AppKey)), req.Id)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_DEFAULT)
		return
	}
	if time.Now().UnixMilli()-record.CreatedTime > 10*60*1000 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_QRCODE_EXPIRED)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	if record.Status == dbModels.QrCodeRecordStatus_OK {
		userId := record.UserId
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
		responses.SuccessHttpResp(ctx, &models.LoginUserResp{
			UserId:        userId,
			NickName:      user.Nickname,
			Avatar:        user.UserPortrait,
			Authorization: services.GenerateToken(appkey, userId),
			ImToken:       resp.Token,
		})
	} else if record.Status == dbModels.QrCodeRecordStatus_Default {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_CONTINUE)
		return
	}
}

func ConfirmQrCode(ctx *gin.Context) {
	req := &models.QrCode{}
	if err := ctx.BindJSON(req); err != nil || req.Id == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	userId := ctx.GetString(string(ctxs.CtxKey_RequesterId))
	storage := storages.NewQrCodeRecordStorage()
	err := storage.UpdateStatus(appkey, req.Id, dbModels.QrCodeRecordStatus_OK, userId)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_DEFAULT)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
