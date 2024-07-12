package apis

import (
	"appserver/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func QryFriends(ctx *gin.Context) {
	userId := ctx.Query("user_id")
	startId := ctx.Query("start_id")
	count := 20
	var err error
	countStr := ctx.Query("count")
	if countStr != "" {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			count = 20
		}
	}
	code, friends := services.QryFrineds(userId, startId, count)
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(friends))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}

func AddFriend(ctx *gin.Context) {
	var req services.Friend
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamErr))
		return
	}
	code := services.AddFriend(req.UserId, req.FriendId)
	ctx.JSON(http.StatusOK, services.GetError(code))
}
