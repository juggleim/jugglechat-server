package apis

import (
	"appserver/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UpdateUser(ctx *gin.Context) {
	var req services.User
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamRequired))
		return
	}
	code := services.UpdateUser(req)
	ctx.JSON(http.StatusOK, services.GetError(code))
}

func SearchByPhone(ctx *gin.Context) {
	var req services.User
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamRequired))
		return
	}
	curUserId := GetCurrentUserId(ctx)
	code, users := services.SearchByPhone(curUserId, req.Phone)
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(users))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}

func QryUserInfo(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	code, user := services.QryUserInfo(GetCurrentUserId(ctx), userId)
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(user))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}
