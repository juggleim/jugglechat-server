package apis

import (
	"appserver/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateGroup(ctx *gin.Context) {
	var req services.Group
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamErr))
		return
	}
	code, grp := services.CreateGroup(GetCurrentUserId(ctx), req)
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(grp))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}

func UpdateGroup(ctx *gin.Context) {
	var req services.Group
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamErr))
		return
	}
	code := services.UpdateGroup(GetCurrentUserId(ctx), req)
	ctx.JSON(http.StatusOK, services.GetError(code))
}

func AddGrpMembers(ctx *gin.Context) {
	var req services.Group
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamErr))
		return
	}
	code := services.AddGroupMembers(GetCurrentUserId(ctx), req)
	ctx.JSON(http.StatusOK, services.GetError(code))
}

func DelGrpMembers(ctx *gin.Context) {
	var req services.Group
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, services.GetError(services.ErrorCode_ParamErr))
		return
	}
	code := services.DelGroupMembers(GetCurrentUserId(ctx), req)
	ctx.JSON(http.StatusOK, services.GetError(code))
}

func QryGroup(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	code, grp := services.QryGroup(groupId)
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(grp))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}

func QryMyGroups(ctx *gin.Context) {
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
	code, grps := services.QryMyGroups(GetCurrentUserId(ctx), startId, int64(count))
	if code == services.ErrorCode_Success {
		ctx.JSON(http.StatusOK, services.SuccessResp(grps))
	} else {
		ctx.JSON(http.StatusOK, services.GetError(code))
	}
}
