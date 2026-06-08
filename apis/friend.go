package apis

import (
	"strconv"

	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/gin-gonic/gin"
)

func QryFriends(ctx *gin.Context) {
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
	code, friends := services.QryFriends(ctxs.ToCtx(ctx), int64(count), offset)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	ret := &models.Friends{
		Items:  []*models.UserObj{},
		Offset: friends.Offset,
	}
	for _, friend := range friends.Items {
		ret.Items = append(ret.Items, &models.UserObj{
			UserId:   friend.UserId,
			Nickname: friend.Nickname,
			Avatar:   friend.Avatar,
			Pinyin:   friend.Pinyin,
			IsFriend: true,
		})
	}
	responses.SuccessHttpResp(ctx, ret)
}

func QryFriendsWithPage(ctx *gin.Context) {
	var err error
	page := 1
	pageStr := ctx.Query("page")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			page = 1
		}
	}

	size := 20
	sizeStr := ctx.Query("size")
	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			size = 20
		}
	}
	orderTag := ctx.Query("order_tag")
	code, friends := services.QryFriendsWithPage(ctxs.ToCtx(ctx), int64(page), int64(size), orderTag)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	ret := &models.Friends{
		Items: []*models.UserObj{},
	}
	for _, friend := range friends.Items {
		ret.Items = append(ret.Items, &models.UserObj{
			UserId:     friend.UserId,
			Nickname:   friend.Nickname,
			Avatar:     friend.Avatar,
			Pinyin:     friend.Pinyin,
			IsFriend:   true,
			FriendInfo: friend.FriendInfo,
		})
	}
	responses.SuccessHttpResp(ctx, ret)
}

func SearchFriends(ctx *gin.Context) {
	req := models.SearchFriendsReq{}
	if err := ctx.BindJSON(&req); err != nil || req.Key == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	if services.CheckApiBlockByVersion(ctxs.ToCtx(ctx)) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_FORBIDDEN)
		return
	}
	code, resp := services.SearchFriends(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func AddFriend(ctx *gin.Context) {
	req := models.Friend{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.AddFriends(ctxs.ToCtx(ctx), &models.FriendIdsReq{
		FriendIds: []string{req.FriendId},
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DelFriend(ctx *gin.Context) {
	req := models.FriendIds{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DelFriends(ctxs.ToCtx(ctx), &models.FriendIdsReq{
		FriendIds: req.FriendIds,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func ApplyFriend(ctx *gin.Context) {
	req := models.ApplyFriend{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	if services.CheckApiBlockByVersion(ctxs.ToCtx(ctx)) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_FORBIDDEN)
		return
	}
	code := services.ApplyFriend(ctxs.ToCtx(ctx), &models.ApplyFriend{
		FriendId: req.FriendId,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func ConfirmFriend(ctx *gin.Context) {
	req := models.ConfirmFriend{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.ConfirmFriend(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func MyFriendApplications(ctx *gin.Context) {
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryMyFriendApplications(ctxs.ToCtx(ctx), start, int32(count), int32(order))
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func MyPendingFriendApplications(ctx *gin.Context) {
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryMyPendingFriendApplications(ctxs.ToCtx(ctx), start, int32(count), int32(order))
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func FriendApplications(ctx *gin.Context) {
	startTimeStr := ctx.Query("start")
	start, err := utils.String2Int64(startTimeStr)
	if err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	countStr := ctx.Query("count")
	count, err := utils.String2Int64(countStr)
	if err != nil {
		count = 20
	} else {
		if count <= 0 || count > 50 {
			count = 20
		}
	}
	orderStr := ctx.Query("order")
	order, err := utils.String2Int64(orderStr)
	if err != nil || order > 1 || order < 0 {
		order = 0
	}
	code, resp := services.QryFriendApplications(ctxs.ToCtx(ctx), start, count, int32(order))
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func SetFriendDisplayName(ctx *gin.Context) {
	req := models.SetFriendDisplayNameReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetFriendDisplayName(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
