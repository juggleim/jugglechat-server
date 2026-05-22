package services

import (
	"context"
	"fmt"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/tools"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
)

func TestGroup(ctx context.Context) errs.IMErrorCode {
	fmt.Println("xxxxxxxxxx")
	return errs.IMErrorCode_SUCCESS
}

func QryGroupInfo(ctx context.Context, groupId string) (errs.IMErrorCode, *apimodels.GrpInfo) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	grpStorage := storages.NewGroupStorage()
	grpInfo, err := grpStorage.FindById(appkey, groupId)
	if err != nil || grpInfo == nil {
		return errs.IMErrorCode_APP_DEFAULT, nil
	}
	grpMemberStorage := storages.NewGroupMemberStorage()
	memberCount := grpMemberStorage.CountByGroup(appkey, groupId)
	ret := &apimodels.GrpInfo{
		GroupId:       grpInfo.GroupId,
		GroupName:     grpInfo.GroupName,
		GroupPortrait: grpInfo.GroupPortrait,
		Members:       []*apimodels.GroupMemberInfo{},
		MemberCount:   int32(memberCount),
		Owner:         &apimodels.GroupMemberInfo{},
		GroupManagement: &apimodels.GroupManagement{
			GroupMute:          grpInfo.IsMute,
			MaxAdminCount:      10,
			GroupHisMsgVisible: 1,

			GroupEditMsgRight:     utils.IntPtr(7),
			GroupAddMemberRight:   utils.IntPtr(7),
			GroupMentionAllRight:  utils.IntPtr(7),
			GroupTopMsgRight:      utils.IntPtr(7),
			GroupSendMsgRight:     utils.IntPtr(7),
			GroupSetMsgLifeRight:  utils.IntPtr(7),
			GroupApplyFriendRight: utils.IntPtr(7),
		},
	}
	isMember := false
	//check is member
	member, err := grpMemberStorage.Find(appkey, groupId, requestId)
	if err == nil && member != nil {
		isMember = true
		ret.GrpDisplayName = member.GrpDisplayName
	}
	if !isMember {
		ret.MyRole = apimodels.GrpMemberRole_GrpNotMember
		return errs.IMErrorCode_SUCCESS, ret
	}
	//grp setting
	grpExtStorage := storages.NewGroupExtStorage()
	exts, err := grpExtStorage.QryExtFields(appkey, groupId)
	if err == nil {
		for _, ext := range exts {
			if ext.ItemKey == apimodels.AttItemKey_GrpVerifyType {
				verifyType := utils.ToInt(ext.ItemValue)
				ret.GroupManagement.GroupVerifyType = verifyType
			} else if ext.ItemKey == apimodels.AttItemKey_HideGrpMsg {
				hidGrpMsg := utils.ToInt(ext.ItemValue)
				var visible int = 0
				if hidGrpMsg > 0 {
					visible = 0
				} else {
					visible = 1
				}
				ret.GroupManagement.GroupHisMsgVisible = *utils.IntPtr(visible)
			} else if ext.ItemKey == apimodels.AttItemKey_GrpEditMsgRight {
				editMsgRight := utils.ToInt(ext.ItemValue)
				ret.GroupManagement.GroupEditMsgRight = utils.IntPtr(editMsgRight)
			} else if ext.ItemKey == apimodels.AttItemKey_AddMemberRight {
				ret.GroupManagement.GroupAddMemberRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_MentionAllRight {
				ret.GroupManagement.GroupMentionAllRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_TopMsgRight {
				ret.GroupManagement.GroupTopMsgRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_SendMsgRight {
				ret.GroupManagement.GroupSendMsgRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_SetMsgLifeRight {
				ret.GroupManagement.GroupSetMsgLifeRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_ApplyFriendRight {
				ret.GroupManagement.GroupApplyFriendRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			}
		}
	}
	//grp administrator
	administrators := map[string]bool{}
	grpAdminStorage := storages.NewGroupAdminStorage()
	admins, err := grpAdminStorage.QryAdmins(appkey, groupId)
	if err == nil {
		for _, admin := range admins {
			administrators[admin.AdminId] = true
		}
	}
	// my role
	myRole := apimodels.GrpMemberRole_GrpMember // 0: 群成员；1:群主；2:群管理员；3:非群成员；
	if requestId == grpInfo.CreatorId {
		myRole = apimodels.GrpMemberRole_GrpCreator
	} else if _, exist := administrators[requestId]; exist {
		myRole = apimodels.GrpMemberRole_GrpAdmin
	}
	ret.MyRole = myRole
	//owner
	if grpInfo.CreatorId != "" {
		ownerUser := GetUser(ctx, grpInfo.CreatorId)
		if ownerUser != nil {
			ret.Owner = &apimodels.GroupMemberInfo{
				UserId:     ownerUser.UserId,
				Nickname:   ownerUser.Nickname,
				Avatar:     ownerUser.Avatar,
				MemberType: ownerUser.UserType,
			}
		}
	}
	//top members
	topMembers, err := grpMemberStorage.QueryMembers(appkey, requestId, groupId, 0, 20)
	if err == nil && len(topMembers) > 0 {
		for _, member := range topMembers {
			role := apimodels.GrpMemberRole_GrpMember
			if member.MemberId == grpInfo.CreatorId {
				role = apimodels.GrpMemberRole_GrpCreator
			} else if _, exist := administrators[member.MemberId]; exist {
				role = apimodels.GrpMemberRole_GrpAdmin
			}
			friendInfo := &apimodels.FriendInfo{
				IsFriend: false,
			}
			if member.MemberFriendInfo != nil {
				friendInfo.IsFriend = member.MemberFriendInfo.IsFriend
				friendInfo.DisplayName = member.MemberFriendInfo.DisplayName
			}
			ret.Members = append(ret.Members, &apimodels.GroupMemberInfo{
				UserId:     member.MemberId,
				Role:       role,
				Nickname:   member.Nickname,
				Avatar:     member.UserPortrait,
				MemberType: member.MemberType,
				IsMute:     member.IsMute,

				FriendInfo: friendInfo,
			})
			ret.MemberOffset, _ = utils.EncodeInt(member.ID)
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func GetGroupInfo(ctx context.Context, groupId string) *apimodels.GrpInfo {
	if groupId == "" {
		return nil
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	ret := &apimodels.GrpInfo{
		GroupId: groupId,
	}
	grpStorage := storages.NewGroupStorage()
	grpInfo, err := grpStorage.FindById(appkey, groupId)
	if err == nil && grpInfo != nil {
		ret.GroupName = grpInfo.GroupName
		ret.GroupPortrait = grpInfo.GroupPortrait
	}
	return ret
}

func GetGroupSettings(ctx context.Context, groupId string) *apimodels.GroupManagement {
	ret := &apimodels.GroupManagement{
		GroupEditMsgRight:     utils.IntPtr(7),
		GroupAddMemberRight:   utils.IntPtr(7),
		GroupMentionAllRight:  utils.IntPtr(7),
		GroupTopMsgRight:      utils.IntPtr(7),
		GroupSendMsgRight:     utils.IntPtr(7),
		GroupSetMsgLifeRight:  utils.IntPtr(7),
		GroupApplyFriendRight: utils.IntPtr(7),
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	grpExtStorage := storages.NewGroupExtStorage()
	exts, err := grpExtStorage.QryExtFields(appkey, groupId)
	if err == nil {
		for _, ext := range exts {
			if ext.ItemKey == apimodels.AttItemKey_GrpVerifyType {
				verifyType := utils.ToInt(ext.ItemValue)
				ret.GroupVerifyType = verifyType
			} else if ext.ItemKey == apimodels.AttItemKey_HideGrpMsg {
				hidGrpMsg := utils.ToInt(ext.ItemValue)
				var visible int = 0
				if hidGrpMsg > 0 {
					visible = 0
				} else {
					visible = 1
				}
				ret.GroupHisMsgVisible = *utils.IntPtr(visible)
			} else if ext.ItemKey == apimodels.AttItemKey_GrpEditMsgRight {
				editMsgRight := utils.ToInt(ext.ItemValue)
				ret.GroupEditMsgRight = utils.IntPtr(editMsgRight)
			} else if ext.ItemKey == apimodels.AttItemKey_AddMemberRight {
				ret.GroupAddMemberRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_MentionAllRight {
				ret.GroupMentionAllRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_TopMsgRight {
				ret.GroupTopMsgRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_SendMsgRight {
				ret.GroupSendMsgRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_SetMsgLifeRight {
				ret.GroupSetMsgLifeRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			} else if ext.ItemKey == apimodels.AttItemKey_ApplyFriendRight {
				ret.GroupApplyFriendRight = utils.IntPtr(utils.ToInt(ext.ItemValue))
			}
		}
	}
	return ret
}

func CheckGroupMembers(ctx context.Context, req *apimodels.CheckGroupMembersReq) (errs.IMErrorCode, *apimodels.CheckGroupMembersResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	ret := &apimodels.CheckGroupMembersResp{
		GroupId:        req.GroupId,
		MemberExistMap: map[string]bool{},
	}
	for _, memberId := range req.MemberIds {
		ret.MemberExistMap[memberId] = false
	}
	storage := storages.NewGroupMemberStorage()
	members, err := storage.FindByMemberIds(appkey, req.GroupId, req.MemberIds)
	if err == nil {
		for _, member := range members {
			ret.MemberExistMap[member.MemberId] = true
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SearchGroupMembers(ctx context.Context, req *apimodels.SearchGroupMembersReq) (errs.IMErrorCode, *apimodels.GroupMemberInfos) {
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	groupId := req.GroupId
	var startId int64 = 0
	if req.Offset != "" {
		id, err := utils.DecodeInt(req.Offset)
		if err == nil && id > 0 {
			startId = id
		}
	}
	var limit int64 = 100
	if req.Limit > 0 {
		limit = req.Limit
	}
	ret := &apimodels.GroupMemberInfos{
		Items: []*apimodels.GroupMemberInfo{},
	}
	storage := storages.NewGroupMemberStorage()
	members, err := storage.SearchMembersByName(appkey, userId, groupId, req.Key, startId, limit)
	if err == nil {
		seenMemberIds := map[string]struct{}{}
		for _, member := range members {
			ret.Offset, _ = utils.EncodeInt(member.ID)
			if _, ok := seenMemberIds[member.MemberId]; ok {
				continue
			}
			seenMemberIds[member.MemberId] = struct{}{}
			friendInfo := &apimodels.FriendInfo{}
			if member.MemberFriendInfo != nil {
				friendInfo.IsFriend = member.MemberFriendInfo.IsFriend
				friendInfo.DisplayName = member.MemberFriendInfo.DisplayName
			}
			ret.Items = append(ret.Items, &apimodels.GroupMemberInfo{
				UserId:     member.MemberId,
				MemberType: member.MemberType,
				Nickname:   member.Nickname,
				Avatar:     member.UserPortrait,
				FriendInfo: friendInfo,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func CreateGroup(ctx context.Context, req *apimodels.GroupMembersReq) (errs.IMErrorCode, *apimodels.GroupInfo) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	grpName := req.GroupName
	//check grpName
	if ok, _ := CheckSensitiveText(ctx, grpName); !ok {
		return errs.IMErrorCode_APP_Sensitive, nil
	}
	grpId := utils.GenerateUUIDShort11()
	if req.GroupId != "" {
		grpId = req.GroupId
	}
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	memberIds := []string{requestId}
	for _, memberId := range req.MemberIds {
		if memberId != requestId {
			memberIds = append(memberIds, memberId)
		}
	}
	storage := storages.NewGroupStorage()
	storage.Create(models.Group{
		GroupId:       grpId,
		GroupName:     req.GroupName,
		GroupPortrait: req.GroupPortrait,
		CreatorId:     requestId,
		AppKey:        appkey,
	})
	memberStorage := storages.NewGroupMemberStorage()
	items := []models.GroupMember{}
	for _, mId := range memberIds {
		items = append(items, models.GroupMember{
			GroupId:  grpId,
			MemberId: mId,
			AppKey:   appkey,
		})
	}
	memberStorage.BatchCreate(items)
	//sync to im server
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		code, _, err := sdk.CreateGroup(juggleimsdk.GroupMembersReq{
			GroupId:       grpId,
			GroupName:     req.GroupName,
			GroupPortrait: req.GroupPortrait,
			MemberIds:     memberIds,
		})
		if err != nil || code != 0 {
			return errs.IMErrorCode(code), nil
		}
	}
	//send notify msg
	targetUsers := []*apimodels.UserObj{}
	for _, memberId := range req.MemberIds {
		targetUsers = append(targetUsers, GetUser(ctx, memberId))
	}
	notify := &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Members:  targetUsers,
		Type:     apimodels.GroupNotifyType_AddMember,
	}
	SendGrpNotify(ctx, grpId, notify)
	// set group confs
	grpExtStorage := storages.NewGroupExtStorage()
	grpExtStorage.Upsert(models.GroupExt{
		GroupId:   grpId,
		ItemKey:   apimodels.AttItemKey_AddMemberRight,
		ItemValue: "3",
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	return errs.IMErrorCode_SUCCESS, &apimodels.GroupInfo{
		GroupId:       grpId,
		GroupName:     req.GroupName,
		GroupPortrait: req.GroupPortrait,
	}
}

func UpdateGroup(ctx context.Context, req *apimodels.GroupInfo) errs.IMErrorCode {
	if ok, _ := CheckSensitiveText(ctx, req.GroupName); !ok {
		return errs.IMErrorCode_APP_Sensitive
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGroupStorage()
	storage.UpdateGrpName(appkey, req.GroupId, req.GroupName, req.GroupPortrait)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.UpdateGroup(juggleimsdk.GroupInfo{
			GroupId:       req.GroupId,
			GroupName:     req.GroupName,
			GroupPortrait: req.GroupPortrait,
		})
	}
	SendGrpNotify(ctx, req.GroupId, &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Name:     req.GroupName,
		Type:     apimodels.GroupNotifyType_Rename,
	})
	return errs.IMErrorCode_SUCCESS
}

func DissolveGroup(ctx context.Context, groupId string) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGroupStorage()
	//check owner
	grp, err := storage.FindById(appkey, groupId)
	if err != nil || grp.CreatorId != userId {
		return errs.IMErrorCode_APP_GROUP_DEFAULT
	}
	storage.Delete(appkey, groupId)
	memberStorage := storages.NewGroupMemberStorage()
	memberStorage.DeleteByGroupId(appkey, groupId)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.DissolveGroup(groupId)
	}
	return errs.IMErrorCode_SUCCESS
}

func QuitGroup(ctx context.Context, groupId string) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	memberStorage := storages.NewGroupMemberStorage()
	memberStorage.BatchDelete(appkey, groupId, []string{requestId})
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.GroupDelMembers(juggleimsdk.GroupMembersReq{
			GroupId:   groupId,
			MemberIds: []string{requestId},
		})
	}
	SendGrpNotify(ctx, groupId, &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Members: []*apimodels.UserObj{
			GetUser(ctx, requestId),
		},
		Type: apimodels.GroupNotifyType_RemoveMember,
	})
	return errs.IMErrorCode_SUCCESS
}

func AddGrpMembers(ctx context.Context, grpMembers *apimodels.GroupMembersReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	memberStorage := storages.NewGroupMemberStorage()
	items := []models.GroupMember{}
	for _, mId := range grpMembers.MemberIds {
		items = append(items, models.GroupMember{
			GroupId:  grpMembers.GroupId,
			MemberId: mId,
			AppKey:   appkey,
		})
	}
	memberStorage.BatchCreate(items)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.GroupAddMembers(juggleimsdk.GroupMembersReq{
			GroupId:       grpMembers.GroupId,
			GroupName:     grpMembers.GroupName,
			GroupPortrait: grpMembers.GroupPortrait,
			MemberIds:     grpMembers.MemberIds,
		})
	}
	//send notify msg
	targetUsers := []*apimodels.UserObj{}
	for _, memberId := range grpMembers.MemberIds {
		targetUsers = append(targetUsers, GetUser(ctx, memberId))
	}
	notify := &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Members:  targetUsers,
		Type:     apimodels.GroupNotifyType_AddMember,
	}
	//send notify msg
	SendGrpNotify(ctx, grpMembers.GroupId, notify)
	return errs.IMErrorCode_SUCCESS
}

func GrpInviteMembers(ctx context.Context, req *apimodels.GroupInviteReq) (errs.IMErrorCode, *apimodels.GroupInviteResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requesterId := ctxs.GetRequesterIdFromCtx(ctx)
	//TODO check operator
	results := &apimodels.GroupInviteResp{
		Results: make(map[string]apimodels.GrpInviteResultReason),
	}
	//TODO check grp member exist
	//check group's setting
	grpSettings := GetGroupSettings(ctx, req.GroupId)
	if grpSettings.GroupVerifyType == apimodels.GrpVerifyType_DeclineGroup {
		results.Reason = apimodels.GrpInviteResultReason_InviteDecline
	} else if grpSettings.GroupVerifyType == apimodels.GrpVerifyType_NeedGrpVerify {
		storage := storages.NewGrpApplicationStorage()
		for _, memberId := range req.MemberIds {
			storage.InviteUpsert(models.GrpApplication{
				GroupId:     req.GroupId,
				ApplyType:   models.GrpApplicationType_Invite,
				RecipientId: memberId,
				InviterId:   requesterId,
				ApplyTime:   time.Now().UnixMilli(),
				Status:      models.GrpApplicationStatus_Invite,
				AppKey:      appkey,
			})
		}
		results.Reason = apimodels.GrpInviteResultReason_InviteSendOut
	} else {
		//check user's setting
		directAddMemberIds := []string{}
		for _, memberId := range req.MemberIds {
			reason := apimodels.GrpInviteResultReason_InviteSucc
			mUserSetting := GetUserSettings(ctx, memberId)
			if mUserSetting.GrpVerifyType == apimodels.GrpVerifyType_DeclineGroup {
				reason = apimodels.GrpInviteResultReason_InviteDecline
			} else if mUserSetting.GrpVerifyType == apimodels.GrpVerifyType_NeedGrpVerify {
				storage := storages.NewGrpApplicationStorage()
				storage.InviteUpsert(models.GrpApplication{
					GroupId:     req.GroupId,
					ApplyType:   models.GrpApplicationType_Invite,
					RecipientId: memberId,
					InviterId:   requesterId,
					ApplyTime:   time.Now().UnixMilli(),
					Status:      models.GrpApplicationStatus_Invite,
					AppKey:      appkey,
				})
				reason = apimodels.GrpInviteResultReason_InviteSendOut
			} else if mUserSetting.GrpVerifyType == apimodels.GrpVerifyType_NoNeedGrpVerify {
				directAddMemberIds = append(directAddMemberIds, memberId)
				reason = apimodels.GrpInviteResultReason_InviteSucc
			}
			results.Results[memberId] = reason
		}
		if len(directAddMemberIds) > 0 {
			memberStorage := storages.NewGroupMemberStorage()
			items := []models.GroupMember{}
			for _, mId := range directAddMemberIds {
				items = append(items, models.GroupMember{
					GroupId:  req.GroupId,
					MemberId: mId,
					AppKey:   appkey,
				})
			}
			memberStorage.BatchCreate(items)
			//sync to imserver
			if sdk := imsdk.GetImSdk(appkey); sdk != nil {
				sdk.GroupAddMembers(juggleimsdk.GroupMembersReq{
					GroupId:   req.GroupId,
					MemberIds: directAddMemberIds,
				})
			}
			//send notify msg
			targetUsers := []*apimodels.UserObj{}
			for _, memberId := range directAddMemberIds {
				targetUsers = append(targetUsers, GetUser(ctx, memberId))
			}
			notify := &apimodels.GroupNotify{
				Operator: GetUser(ctx, requesterId),
				Members:  targetUsers,
				Type:     apimodels.GroupNotifyType_AddMember,
			}
			SendGrpNotify(ctx, req.GroupId, notify)
		}
	}
	return errs.IMErrorCode_SUCCESS, results
}

func GroupConfirm(ctx context.Context, req *apimodels.GroupConfirm) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	//TODO check admin
	storage := storages.NewGrpApplicationStorage()
	id, err := utils.DecodeInt(req.ApplicationId)
	if err != nil || id <= 0 {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	application, err := storage.FindById(appkey, id)
	if err != nil || application == nil {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL
	}
	if application.ApplyType == models.GrpApplicationType_Invite {
		if req.IsAgree {
			err := storage.UpdateStatus(id, models.GrpApplicationStatus_AgreeInvite)
			if err == nil {
				memberStorage := storages.NewGroupMemberStorage()
				memberStorage.BatchCreate([]models.GroupMember{
					{
						GroupId:  application.GroupId,
						MemberId: application.RecipientId,
						AppKey:   appkey,
					},
				})
				//sync to imserver
				if sdk := imsdk.GetImSdk(appkey); sdk != nil {
					sdk.GroupAddMembers(juggleimsdk.GroupMembersReq{
						GroupId:   application.GroupId,
						MemberIds: []string{application.RecipientId},
					})
				}
				//send notify msg
				targetUsers := []*apimodels.UserObj{
					GetUser(ctx, application.RecipientId),
				}
				notify := &apimodels.GroupNotify{
					Operator: GetUser(ctx, application.InviterId),
					Members:  targetUsers,
					Type:     apimodels.GroupNotifyType_AddMember,
				}
				SendGrpNotify(ctx, application.GroupId, notify)
			}
		} else {
			storage.UpdateStatus(id, models.GrpApplicationStatus_DeclineInvite)
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func GrpJoinApply(ctx context.Context, req *apimodels.GroupInviteReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	groupId := req.GroupId

	//check grp member exists
	memberStorage := storages.NewGroupMemberStorage()
	member, err := memberStorage.Find(appkey, groupId, userId)
	if err == nil && member != nil {
		return errs.IMErrorCode_APP_GROUP_MEMBEREXISTED
	}

	//add group
	memberStorage.Create(models.GroupMember{
		GroupId:  groupId,
		MemberId: userId,
		AppKey:   appkey,
	})
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.GroupAddMembers(juggleimsdk.GroupMembersReq{
			GroupId:   groupId,
			MemberIds: []string{userId},
		})
	}
	//send notify msg
	notify := &apimodels.GroupNotify{
		Operator: GetUser(ctx, userId),
		Type:     apimodels.GroupNotifyType_Join,
	}
	SendGrpNotify(ctx, groupId, notify)
	return errs.IMErrorCode_SUCCESS
}

func DelGrpMembers(ctx context.Context, req *apimodels.GroupMembersReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)

	memberStorage := storages.NewGroupMemberStorage()
	memberStorage.BatchDelete(appkey, req.GroupId, req.MemberIds)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.GroupDelMembers(juggleimsdk.GroupMembersReq{
			GroupId:   req.GroupId,
			MemberIds: req.MemberIds,
		})
	}
	//send notify msg
	targetUsers := []*apimodels.UserObj{}
	for _, memberId := range req.MemberIds {
		targetUsers = append(targetUsers, GetUser(ctx, memberId))
	}
	SendGrpNotify(ctx, req.GroupId, &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Members:  targetUsers,
		Type:     apimodels.GroupNotifyType_RemoveMember,
	})
	return errs.IMErrorCode_SUCCESS
}

func QueryGrpMembers(ctx context.Context, groupId string, limit int64, offset string) (errs.IMErrorCode, *apimodels.GroupMemberInfos) {
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGroupMemberStorage()
	var startId int64 = 0
	if offset != "" {
		id, err := utils.DecodeInt(offset)
		if err == nil && id > 0 {
			startId = id
		}
	}
	ret := &apimodels.GroupMemberInfos{
		Items: []*apimodels.GroupMemberInfo{},
	}
	members, err := storage.QueryMembers(appkey, userId, groupId, startId, limit)
	if err == nil {
		grpStorage := storages.NewGroupStorage()
		grpInfo, err := grpStorage.FindById(appkey, groupId)
		if err != nil || grpInfo == nil {
			return errs.IMErrorCode_APP_DEFAULT, nil
		}
		//grp administrator
		administrators := map[string]bool{}
		grpAdminStorage := storages.NewGroupAdminStorage()
		admins, err := grpAdminStorage.QryAdmins(appkey, groupId)
		if err == nil {
			for _, admin := range admins {
				administrators[admin.AdminId] = true
			}
		}
		for _, member := range members {
			friendInfo := &apimodels.FriendInfo{}
			if member.MemberFriendInfo != nil {
				friendInfo.IsFriend = member.MemberFriendInfo.IsFriend
				friendInfo.DisplayName = member.MemberFriendInfo.DisplayName
			}
			ret.Offset, _ = utils.EncodeInt(member.ID)
			role := apimodels.GrpMemberRole_GrpMember
			if member.MemberId == grpInfo.CreatorId {
				role = apimodels.GrpMemberRole_GrpCreator
			} else if _, exist := administrators[member.MemberId]; exist {
				role = apimodels.GrpMemberRole_GrpAdmin
			}
			ret.Items = append(ret.Items, &apimodels.GroupMemberInfo{
				UserId:     member.MemberId,
				MemberType: member.MemberType,
				Nickname:   member.Nickname,
				Avatar:     member.UserPortrait,
				FriendInfo: friendInfo,
				Role:       role,
				IsMute:     member.IsMute,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SetGrpAnnouncement(ctx context.Context, req *apimodels.GroupAnnouncement) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGroupExtStorage()
	storage.Upsert(models.GroupExt{
		GroupId:   req.GroupId,
		ItemKey:   apimodels.AttItemKey_GrpAnnouncement,
		ItemValue: req.Content,
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	if req.Content != "" {
		//send announce msg
		SendGroupMsg(ctx, requestId, req.GroupId, "jg:text", map[string]string{
			"content": "{all}" + req.Content,
		}, &juggleimsdk.MentionInfo{
			MentionType: juggleimsdk.MentionType_All,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func GetGrpAnnouncement(ctx context.Context, groupId string) (errs.IMErrorCode, *apimodels.GroupAnnouncement) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGroupExtStorage()
	ret := &apimodels.GroupAnnouncement{
		GroupId: groupId,
	}
	ext, err := storage.Find(appkey, groupId, apimodels.AttItemKey_GrpAnnouncement)
	if err == nil && ext != nil {
		ret.Content = ext.ItemValue
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func ChgGroupOwner(ctx context.Context, req *apimodels.GroupOwnerChgReq) errs.IMErrorCode {
	//TODO check right
	storage := storages.NewGroupStorage()
	storage.UpdateCreatorId(ctxs.GetAppKeyFromCtx(ctx), req.GroupId, req.OwnerId)
	//send notify
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	notify := &apimodels.GroupNotify{
		Operator: GetUser(ctx, requestId),
		Members: []*apimodels.UserObj{
			GetUser(ctx, req.OwnerId),
		},
		Type: apimodels.GroupNotifyType_ChgOwner,
	}
	SendGrpNotify(ctx, req.GroupId, notify)
	return errs.IMErrorCode_SUCCESS
}

func SetGroupMute(ctx context.Context, req *apimodels.SetGroupMuteReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	//TODO check right
	storage := storages.NewGroupStorage()
	storage.UpdateGroupMuteStatus(appkey, req.GroupId, req.IsMute)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.SetGroupMute(req.GroupId, int(req.IsMute))
	}
	return errs.IMErrorCode_SUCCESS
}

func SetGroupMembersMute(ctx context.Context, req *apimodels.SetGroupMemberMuteReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	//TODO check right
	storage := storages.NewGroupMemberStorage()
	storage.UpdateMute(appkey, req.GroupId, int(req.IsMute), req.MemberIds, 0)
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.SetGroupMembersMute(req.GroupId, int(req.IsMute), req.MemberIds)
	}
	return errs.IMErrorCode_SUCCESS
}

func SetGroupVerifyType(ctx context.Context, req *apimodels.SetGroupVerifyTypeReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGroupExtStorage()
	storage.Upsert(models.GroupExt{
		GroupId:   req.GroupId,
		ItemKey:   apimodels.AttItemKey_GrpVerifyType,
		ItemValue: utils.Int2String(int64(req.VerifyType)),
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	return errs.IMErrorCode_SUCCESS
}

func SetGroupHisMsgVisible(ctx context.Context, req *apimodels.SetGroupHisMsgVisibleReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	//TODO check right
	visible := req.GroupHisMsgVisible
	var hideGrpMsg int64 = 1
	if visible > 0 {
		hideGrpMsg = 0
	} else {
		hideGrpMsg = 1
	}
	storage := storages.NewGroupExtStorage()
	storage.Upsert(models.GroupExt{
		GroupId:   req.GroupId,
		ItemKey:   apimodels.AttItemKey_HideGrpMsg,
		ItemValue: tools.Int2String(hideGrpMsg),
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	//sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.SetGroupSettings(juggleimsdk.SetGroupSettingReq{})
		sdk.SetGroupSettings(juggleimsdk.SetGroupSettingReq{
			GroupId: req.GroupId,
			Settings: &juggleimsdk.GroupSettings{
				HideGrpMsg: tools.Int64Ptr(hideGrpMsg),
			},
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func SetGroupManagementConfs(ctx context.Context, req *apimodels.GroupManagement) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	//TODO check right
	items := []models.GroupExt{}
	if req.GroupEditMsgRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_GrpEditMsgRight,
			ItemValue: utils.Int2String(int64(*req.GroupEditMsgRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupAddMemberRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_AddMemberRight,
			ItemValue: utils.Int2String(int64(*req.GroupAddMemberRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupMentionAllRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_MentionAllRight,
			ItemValue: utils.Int2String(int64(*req.GroupMentionAllRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupTopMsgRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_TopMsgRight,
			ItemValue: utils.Int2String(int64(*req.GroupTopMsgRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupSendMsgRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_SendMsgRight,
			ItemValue: utils.Int2String(int64(*req.GroupSendMsgRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupSetMsgLifeRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_SetMsgLifeRight,
			ItemValue: utils.Int2String(int64(*req.GroupSetMsgLifeRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if req.GroupApplyFriendRight != nil {
		items = append(items, models.GroupExt{
			GroupId:   req.GroupId,
			ItemKey:   apimodels.AttItemKey_ApplyFriendRight,
			ItemValue: utils.Int2String(int64(*req.GroupApplyFriendRight)),
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
	}
	if len(items) > 0 {
		storage := storages.NewGroupExtStorage()
		err := storage.BatchUpsert(items)
		if err != nil {
			return errs.IMErrorCode_APP_INTERNAL_TIMEOUT
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func AddGroupAdministrators(ctx context.Context, req *apimodels.GroupAdministratorsReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGroupAdminStorage()
	for _, adminId := range req.AdminIds {
		storage.Upsert(models.GroupAdmin{
			GroupId: req.GroupId,
			AdminId: adminId,
			AppKey:  appkey,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func DelGroupAdministrators(ctx context.Context, req *apimodels.GroupAdministratorsReq) errs.IMErrorCode {
	storage := storages.NewGroupAdminStorage()
	storage.BatchDel(ctxs.GetAppKeyFromCtx(ctx), req.GroupId, req.AdminIds)
	return errs.IMErrorCode_SUCCESS
}

func QryGroupAdministrators(ctx context.Context, groupId string) (errs.IMErrorCode, *apimodels.GroupAdministratorsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requesterId := ctxs.GetRequesterIdFromCtx(ctx)
	ret := &apimodels.GroupAdministratorsResp{
		GroupId: groupId,
		Items:   []*apimodels.GroupMemberInfo{},
	}
	storage := storages.NewGroupAdminStorage()
	admins, err := storage.QryAdmins(appkey, groupId)
	if err == nil {
		friendMap := map[string]*models.FriendInfo{}
		adminIds := []string{}
		for _, admin := range admins {
			adminIds = append(adminIds, admin.AdminId)
		}
		if len(adminIds) > 0 {
			friendStorage := storages.NewFriendRelStorage()
			friendRels, ferr := friendStorage.QueryFriendRelsByFriendIds(appkey, requesterId, adminIds)
			if ferr == nil {
				for _, rel := range friendRels {
					friendMap[rel.FriendId] = &models.FriendInfo{
						IsFriend:    true,
						DisplayName: rel.DisplayName,
					}
				}
			}
		}
		for _, admin := range admins {
			friendInfo := &apimodels.FriendInfo{}
			if friend, exist := friendMap[admin.AdminId]; exist {
				friendInfo.IsFriend = friend.IsFriend
				friendInfo.DisplayName = friend.DisplayName
			}
			ret.Items = append(ret.Items, &apimodels.GroupMemberInfo{
				UserId:         admin.AdminId,
				Nickname:       admin.Nickname,
				Avatar:         admin.UserPortrait,
				GrpDisplayName: admin.GrpDisplayName,
				MemberType:     admin.UserType,
				Role:           apimodels.GrpMemberRole_GrpAdmin,
				FriendInfo:     friendInfo,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SetGrpDisplayName(ctx context.Context, req *apimodels.SetGroupDisplayNameReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGroupMemberStorage()
	storage.UpdateGrpDisplayName(appkey, req.GroupId, userId, req.GrpDisplayName)
	return errs.IMErrorCode_SUCCESS
}

func QryMyGrpApplications(ctx context.Context, startTime int64, count int32, order int32, groupId string) (errs.IMErrorCode, *apimodels.QryGrpApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGrpApplicationStorage()
	ret := &apimodels.QryGrpApplicationsResp{
		Items: []*apimodels.GrpApplicationItem{},
	}
	applications, err := storage.QueryMyGrpApplications(appkey, userId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.GrpApplicationItem{
				GrpInfo: &apimodels.GrpInfo{
					GroupId: application.GroupId,
				},
				ApplyType: int32(application.ApplyType),
				Operator: &apimodels.UserObj{
					UserId: application.OperatorId,
				},
				ApplyTime: application.ApplyTime,
				Status:    int32(application.Status),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryMyPendingGrpInvitations(ctx context.Context, startTime int64, count int32, order int32, groupId string) (errs.IMErrorCode, *apimodels.QryGrpApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewGrpApplicationStorage()
	ret := &apimodels.QryGrpApplicationsResp{
		Items: []*apimodels.GrpApplicationItem{},
	}
	applications, err := storage.QueryMyPendingGrpInvitations(appkey, userId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.GrpApplicationItem{
				GrpInfo: &apimodels.GrpInfo{
					GroupId: application.GroupId,
				},
				ApplyType: int32(application.ApplyType),
				Inviter: &apimodels.UserObj{
					UserId: application.InviterId,
				},
				ApplyTime: application.ApplyTime,
				Status:    int32(application.Status),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryGrpInvitations(ctx context.Context, startTime int64, count int32, order int32, groupId string) (errs.IMErrorCode, *apimodels.QryGrpApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGrpApplicationStorage()
	ret := &apimodels.QryGrpApplicationsResp{
		Items: []*apimodels.GrpApplicationItem{},
	}
	applications, err := storage.QueryGrpInvitations(appkey, groupId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.GrpApplicationItem{
				GrpInfo: &apimodels.GrpInfo{
					GroupId: application.GroupId,
				},
				ApplyType: int32(application.ApplyType),
				Recipient: &apimodels.UserObj{
					UserId: application.RecipientId,
				},
				Inviter: &apimodels.UserObj{
					UserId: application.InviterId,
				},
				ApplyTime: application.ApplyTime,
				Status:    int32(application.Status),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryGrpPendingApplications(ctx context.Context, startTime int64, count int32, order int32, groupId string) (errs.IMErrorCode, *apimodels.QryGrpApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGrpApplicationStorage()
	ret := &apimodels.QryGrpApplicationsResp{
		Items: []*apimodels.GrpApplicationItem{},
	}
	applications, err := storage.QueryGrpPendingApplications(appkey, groupId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.GrpApplicationItem{
				GrpInfo: &apimodels.GrpInfo{
					GroupId: application.GroupId,
				},
				ApplyType: int32(application.ApplyType),
				Sponsor: &apimodels.UserObj{
					UserId: application.SponsorId,
				},
				Operator: &apimodels.UserObj{
					UserId: application.OperatorId,
				},
				ApplyTime: application.ApplyTime,
				Status:    int32(application.Status),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryGrpApplications(ctx context.Context, startTime int64, count int32, order int32, groupId string) (errs.IMErrorCode, *apimodels.QryGrpApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewGrpApplicationStorage()
	ret := &apimodels.QryGrpApplicationsResp{
		Items: []*apimodels.GrpApplicationItem{},
	}
	applications, err := storage.QueryGrpApplications(appkey, groupId, startTime, int64(count), order > 0)
	if err == nil {
		grpInfo := GetGroupInfo(ctx, groupId)
		for _, application := range applications {
			idStr, _ := utils.EncodeInt(application.ID)

			ret.Items = append(ret.Items, &apimodels.GrpApplicationItem{
				ApplicationId: idStr,
				GrpInfo:       grpInfo,
				ApplyType:     int32(application.ApplyType),
				Sponsor:       GetUser(ctx, application.SponsorId),
				Operator:      GetUser(ctx, application.OperatorId),
				Recipient:     GetUser(ctx, application.RecipientId),
				Inviter:       GetUser(ctx, application.InviterId),
				ApplyTime:     application.ApplyTime,
				Status:        int32(application.Status),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}
