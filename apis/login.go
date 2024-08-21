package apis

import (
	"appserver/services"
	"appserver/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	imsdk "github.com/juggleim/imserver-sdk-go"
)

const (
	Header_Authorization string = "Authorization"
	CtxKey_Session       string = "CtxKey_Session"
	CtxKey_UserId        string = "CtxKey_UserId"
)

func Validate(ctx *gin.Context) {
	session := utils.GenerateUUIDShort11()
	ctx.Set(CtxKey_Session, session)

	urlPath := ctx.Request.URL.Path
	if strings.HasSuffix(urlPath, "/login") || strings.HasSuffix(urlPath, "/sms_login") || strings.HasSuffix(urlPath, "/sms/send") {
		return
	}
	authStr := ctx.Request.Header.Get(Header_Authorization)
	userId, err := services.ValidateAuthorization(authStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, services.GetError(services.ErrorCode_TokenErr))
		ctx.Abort()
		return
	}
	ctx.Set(CtxKey_UserId, userId)
}

func GetCurrentUserId(ctx *gin.Context) string {
	userId := ctx.Value(CtxKey_UserId)
	if userId == nil {
		return ""
	}
	return userId.(string)
}

type SmsLoginReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func SmsSend(ctx *gin.Context) {
	var req SmsLoginReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamRequired))
		return
	}
	succ := services.SmsSend(req.Phone)
	if succ {
		ctx.JSON(http.StatusOK, services.GetSuccess())
	} else {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_SmsSendFail))
	}
}

func SmsLogin(ctx *gin.Context) {
	var req SmsLoginReq
	if err := ctx.BindJSON(&req); err != nil || req.Phone == "" || req.Code == "" {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamRequired))
		return
	}
	succ := services.CheckPhoneSmsCode(req.Phone, req.Code)
	if succ {
		// 入库
		token, u, err := services.RegisterOrLoginBySms(services.User{
			Phone:    req.Phone,
			Nickname: fmt.Sprintf("user%s", services.RandomSms()),
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.Writer.Header().Set("X-Status", strconv.Itoa(u.Status))

		ctx.JSON(http.StatusOK, services.SuccessResp(services.LoginUserResp{
			UserId:        u.UserId,
			Authorization: token,
			NickName:      u.Nickname,
			Avatar:        u.Avatar,
			Status:        u.Status,
			ImToken: services.RegisterImToken(imsdk.User{
				UserId:       u.UserId,
				Nickname:     u.Nickname,
				UserPortrait: u.Avatar,
			}),
		}))
	} else {
		ctx.JSON(http.StatusForbidden, services.GetError(services.ErrorCode_NotLogin))
	}
}
