package apis

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strconv"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
)

func QryUserInfo(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	if userId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, user := services.QryUserInfo(ctxs.ToCtx(ctx), userId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, user)
}

func SetLoginAccount(ctx *gin.Context) {
	req := &models.SetUserAccountReq{}
	if err := ctx.BindJSON(req); err != nil || req.Account == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetUserAccount(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func UpdateUser(ctx *gin.Context) {
	req := &models.UserObj{}
	if err := ctx.BindJSON(req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.UpdateUser(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func UpdatePass(ctx *gin.Context) {
	req := &models.UpdUserPassReq{}
	if err := ctx.BindJSON(req); err != nil || req.UserId == "" || req.Password == "" || req.NewPassword == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.UpdatePass(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func UpdateUserSettings(ctx *gin.Context) {
	req := &models.UserSettings{}
	if err := ctx.BindJSON(req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.UpdateUserSettings(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func SearchUsers(ctx *gin.Context) {
	req := &models.SearchReq{}
	if err := ctx.BindJSON(req); err != nil || (req.Keyword == "" && req.Phone == "") {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	var code errs.IMErrorCode
	var users *models.Users
	if req.Phone != "" {
		code, users = services.SearchByPhone(ctxs.ToCtx(ctx), req.Phone)
	} else {
		code, users = services.SearchByKeyword(ctxs.ToCtx(ctx), req.Keyword)
	}
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, users)
}

func QryUserQrCode(ctx *gin.Context) {
	userId := ctx.GetString(string(ctxs.CtxKey_RequesterId))

	m := map[string]interface{}{
		"action":  "add_friend",
		"user_id": userId,
	}
	buf := bytes.NewBuffer([]byte{})
	qrCode, _ := qr.Encode(utils.ToJson(m), qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 400, 400)
	err := png.Encode(buf, qrCode)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_DEFAULT)
		return
	}
	responses.SuccessHttpResp(ctx, map[string]string{
		"qr_code": base64.StdEncoding.EncodeToString(buf.Bytes()),
	})
}

func SyncConfs(ctx *gin.Context) {
	responses.SuccessHttpResp(ctx, &models.UserConfs{})
}

func BlockUsers(ctx *gin.Context) {
	req := &models.BlockUsersReq{}
	if err := ctx.BindJSON(req); err != nil || len(req.BlockUserIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.BlockUsers(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func UnBlockUsers(ctx *gin.Context) {
	req := &models.BlockUsersReq{}
	if err := ctx.BindJSON(req); err != nil || len(req.BlockUserIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.UnBlockUsers(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryBlockUsers(ctx *gin.Context) {
	offset := ctx.Query("offset")
	count := 20
	var err error
	countStr := ctx.Query("count")
	if countStr != "" {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			count = 20
		}
	}
	code, blockUsers := services.QryBlockUsers(ctxs.ToCtx(ctx), int64(count), offset)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, blockUsers)
}

func BindEmailSendEmail(ctx *gin.Context) {
	req := &models.BindEmailReq{}
	if err := ctx.BindJSON(req); err != nil || req.Email == "" || req.Code == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.BindEmailSendEmail(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func BindEmail(ctx *gin.Context) {
	req := &models.BindEmailReq{}
	if err := ctx.BindJSON(req); err != nil || req.Email == "" || req.Code == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.BindEmail(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func BindPhoneSendSms(ctx *gin.Context) {
	req := &models.BindPhoneReq{}
	if err := ctx.BindJSON(req); err != nil || req.Phone == "" || req.Code == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.BindPhoneSendSms(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func BindPhone(ctx *gin.Context) {
	req := &models.BindPhoneReq{}
	if err := ctx.BindJSON(req); err != nil || req.Phone == "" || req.Code == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.BindPhone(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryUsersOnlineStauts(ctx *gin.Context) {
	req := &models.UserIds{}
	if err := ctx.BindJSON(req); err != nil || len(req.UserIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	appkey := ctx.GetString(string(ctxs.CtxKey_AppKey))
	if appkey == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		return
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk == nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_NOT_EXISTED)
		return
	}
	resp, code, _, err := sdk.QryUserOnlineStatus(req.UserIds)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_INTERNAL_TIMEOUT)
		return
	}
	if code != juggleimsdk.ApiCode(errs.IMErrorCode_SUCCESS) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode(code))
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}
