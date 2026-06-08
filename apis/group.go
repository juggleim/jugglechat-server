package apis

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strconv"

	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
)

func CreateGroup(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	memberIds := req.MemberIds
	if len(memberIds) <= 0 && len(req.GrpMembers) > 0 {
		ids := []string{}
		for _, member := range req.GrpMembers {
			ids = append(ids, member.UserId)
		}
		memberIds = ids
	}
	code, grpInfo := services.CreateGroup(ctxs.ToCtx(ctx), &models.GroupMembersReq{
		GroupName:     req.GroupName,
		GroupPortrait: req.GroupPortrait,
		MemberIds:     memberIds,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, &models.Group{
		GroupId:       grpInfo.GroupId,
		GroupName:     grpInfo.GroupName,
		GroupPortrait: grpInfo.GroupPortrait,
	})
}

func UpdateGroup(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.UpdateGroup(ctxs.ToCtx(ctx), &models.GroupInfo{
		GroupId:       req.GroupId,
		GroupName:     req.GroupName,
		GroupPortrait: req.GroupPortrait,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DissolveGroup(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DissolveGroup(ctxs.ToCtx(ctx), req.GroupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QuitGroup(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.QuitGroup(ctxs.ToCtx(ctx), req.GroupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func AddGrpMembers(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	memberIds := []string{}
	for _, user := range req.GrpMembers {
		memberIds = append(memberIds, user.UserId)
	}
	code := services.AddGrpMembers(ctxs.ToCtx(ctx), &models.GroupMembersReq{
		GroupId:   req.GroupId,
		MemberIds: memberIds,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func DelGrpMembers(ctx *gin.Context) {
	req := models.Group{}
	if err := ctx.BindJSON(&req); err != nil || req.GroupId == "" || len(req.MemberIds) <= 0 {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.DelGrpMembers(ctxs.ToCtx(ctx), &models.GroupMembersReq{
		GroupId:   req.GroupId,
		MemberIds: req.MemberIds,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryGroupInfo(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	if groupId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	rpcCtx := ctxs.ToCtx(ctx)
	code, grpInfo := services.QryGroupInfo(rpcCtx, groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, grpInfo)
}

func QryMyGroups(ctx *gin.Context) {
	offset := ctx.Query("offset")
	count := 20
	countStr := ctx.Query("count")
	var err error
	if countStr != "" {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			count = 20
		}
	}
	code, grps := services.QueryMyGroups(ctxs.ToCtx(ctx), int64(count), offset)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	ret := &models.Groups{
		Offset: grps.Offset,
		Items:  []*models.Group{},
	}
	for _, grp := range grps.Items {
		ret.Items = append(ret.Items, &models.Group{
			GroupId:       grp.GroupId,
			GroupName:     grp.GroupName,
			GroupPortrait: grp.GroupPortrait,
		})
	}
	responses.SuccessHttpResp(ctx, ret)
}

func SearchMyGroups(ctx *gin.Context) {
	req := &models.SearchReq{}
	if err := ctx.BindJSON(req); err != nil || req.Keyword == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	code, grps := services.SearchMyGroups(ctxs.ToCtx(ctx), req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	ret := &models.Groups{
		Offset: grps.Offset,
		Items:  []*models.Group{},
	}
	for _, grp := range grps.Items {
		ret.Items = append(ret.Items, &models.Group{
			GroupId:       grp.GroupId,
			GroupName:     grp.GroupName,
			GroupPortrait: grp.GroupPortrait,
		})
	}
	responses.SuccessHttpResp(ctx, ret)
}

func QryGrpMembers(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	offset := ctx.Query("offset")
	limit := 100
	limitStr := ctx.Query("limit")
	var err error
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			limit = 100
		}
	}
	code, members := services.QueryGrpMembers(ctxs.ToCtx(ctx), groupId, int64(limit), offset)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, members)
}

func CheckGroupMembers(ctx *gin.Context) {
	req := models.CheckGroupMembersReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code, resp := services.CheckGroupMembers(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func SearchGroupMembers(ctx *gin.Context) {
	req := models.SearchGroupMembersReq{}
	if err := ctx.BindJSON(&req); err != nil || req.GroupId == "" || req.Key == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	if services.CheckApiBlockByVersion(ctxs.ToCtx(ctx)) {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_FORBIDDEN)
		return
	}
	code, resp := services.SearchGroupMembers(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, resp)
}

func SetGrpAnnouncement(ctx *gin.Context) {
	req := models.GroupAnnouncement{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGrpAnnouncement(ctxs.ToCtx(ctx), &models.GroupAnnouncement{
		GroupId: req.GroupId,
		Content: req.Content,
	})
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func GetGrpAnnouncement(ctx *gin.Context) {
	groupId := ctx.Query("group_id")
	code, grpAnnounce := services.GetGrpAnnouncement(ctxs.ToCtx(ctx), groupId)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, &models.GroupAnnouncement{
		GroupId: grpAnnounce.GroupId,
		Content: grpAnnounce.Content,
	})
}

func SetGrpDisplayName(ctx *gin.Context) {
	req := models.SetGroupDisplayNameReq{}
	if err := ctx.BindJSON(&req); err != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetGrpDisplayName(ctxs.ToCtx(ctx), &req)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}

func QryGrpQrCode(ctx *gin.Context) {
	grpId := ctx.Query("group_id")
	if grpId == "" {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	userId := ctx.GetString(string(ctxs.CtxKey_RequesterId))

	m := map[string]interface{}{
		"action":   "join_group",
		"group_id": grpId,
		"user_id":  userId,
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
