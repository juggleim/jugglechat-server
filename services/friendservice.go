package services

import (
	"context"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
)

func QryFriends(ctx context.Context, limit int64, offset string) (errs.IMErrorCode, *apimodels.Users) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendRelStorage()
	var startId int64 = 0
	if offset != "" {
		startId, _ = utils.DecodeInt(offset)
	}
	ret := &apimodels.Users{
		Items:  []*apimodels.UserObj{},
		Offset: "",
	}
	rels, err := storage.QueryFriendRels(appkey, userId, startId, limit)
	if err == nil {
		uIds := []string{}
		for _, rel := range rels {
			ret.Offset, _ = utils.EncodeInt(rel.ID)
			uIds = append(uIds, rel.FriendId)
			ret.Items = append(ret.Items, &apimodels.UserObj{
				UserId: rel.FriendId,
				Pinyin: rel.OrderTag,
			})
		}
		userStorage := storages.NewUserStorage()
		userMap, err := userStorage.FindByUserIds(appkey, uIds)
		if err == nil {
			for _, user := range ret.Items {
				if u, exist := userMap[user.UserId]; exist {
					user.Nickname = u.Nickname
					user.Avatar = u.UserPortrait
					user.UserType = u.UserType
					user.Pinyin = u.Pinyin
				}
			}
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryFriendsWithPage(ctx context.Context, page, size int64, orderTag string) (errs.IMErrorCode, *apimodels.Users) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendRelStorage()
	users, err := storage.QueryFriendRelsWithPage(appkey, userId, orderTag, page, size)
	ret := &apimodels.Users{
		Items: []*apimodels.UserObj{},
	}
	if err == nil {
		for _, user := range users {
			friendInfo := &apimodels.FriendInfo{}
			if user.FriendInfo != nil {
				friendInfo.IsFriend = user.FriendInfo.IsFriend
				friendInfo.DisplayName = user.FriendInfo.DisplayName
			}
			ret.Items = append(ret.Items, &apimodels.UserObj{
				UserId:     user.UserId,
				Pinyin:     user.Pinyin,
				Nickname:   user.Nickname,
				Avatar:     user.UserPortrait,
				UserType:   user.UserType,
				FriendInfo: friendInfo,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SearchFriends(ctx context.Context, req *apimodels.SearchFriendsReq) (errs.IMErrorCode, *apimodels.Users) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	ret := &apimodels.Users{
		Items: []*apimodels.UserObj{},
	}
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
	storage := storages.NewFriendRelStorage()
	users, err := storage.SearchFriendsByName(appkey, userId, req.Key, startId, limit)
	if err == nil {
		for _, u := range users {
			ret.Offset, _ = utils.EncodeInt(u.ID)
			friendInfo := &apimodels.FriendInfo{}
			if u.FriendInfo != nil {
				friendInfo.IsFriend = u.FriendInfo.IsFriend
				friendInfo.DisplayName = u.FriendInfo.DisplayName
			}
			ret.Items = append(ret.Items, &apimodels.UserObj{
				UserId:     u.UserId,
				Nickname:   u.Nickname,
				Avatar:     u.UserPortrait,
				UserType:   u.UserType,
				FriendInfo: friendInfo,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func AddFriends(ctx context.Context, req *apimodels.FriendIdsReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendRelStorage()
	for _, friendId := range req.FriendIds {
		storage.BatchUpsert([]models.FriendRel{
			{
				AppKey:   appkey,
				UserId:   userId,
				FriendId: friendId,
			},
			{
				AppKey:   appkey,
				UserId:   friendId,
				FriendId: userId,
			},
		})
		// sync to imserver
		if sdk := imsdk.GetImSdk(appkey); sdk != nil {
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    userId,
				FriendIds: []string{friendId},
			})
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    friendId,
				FriendIds: []string{userId},
			})
		}
		//send notify msg
		SendFriendNotify(ctx, friendId, &apimodels.FriendNotify{
			Type: 0,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func DelFriends(ctx context.Context, req *apimodels.FriendIdsReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendRelStorage()
	storage.BatchDelete(appkey, userId, req.FriendIds)
	for _, friendId := range req.FriendIds {
		storage.BatchDelete(appkey, friendId, []string{userId})
	}
	// sync to imserver
	if sdk := imsdk.GetImSdk(appkey); sdk != nil {
		sdk.DelFriends(juggleimsdk.FriendIds{
			UserId:    userId,
			FriendIds: req.FriendIds,
		})
		for _, friendId := range req.FriendIds {
			sdk.DelFriends(juggleimsdk.FriendIds{
				UserId:    friendId,
				FriendIds: []string{userId},
			})
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func ApplyFriend(ctx context.Context, req *apimodels.ApplyFriend) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	fStorage := storages.NewFriendRelStorage()
	//check friend relation
	if checkFriend(ctx, req.FriendId, userId) {
		fStorage.Upsert(models.FriendRel{
			AppKey:   appkey,
			UserId:   userId,
			FriendId: req.FriendId,
		})
		// sync to im
		if sdk := imsdk.GetImSdk(appkey); sdk != nil {
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    userId,
				FriendIds: []string{req.FriendId},
			})
		}
		storage := storages.NewFriendApplicationStorage()
		storage.Upsert(models.FriendApplication{
			RecipientId: req.FriendId,
			SponsorId:   userId,
			ApplyTime:   time.Now().UnixMilli(),
			Status:      models.FriendApplicationStatus(models.FriendApplicationStatus_Agree),
			AppKey:      appkey,
		})
		return errs.IMErrorCode_SUCCESS
	}
	friendSettings := GetUserSettings(ctx, req.FriendId)
	if friendSettings.FriendVerifyType == apimodels.FriendVerifyType_DeclineFriend {
		return errs.IMErrorCode_APP_FRIEND_APPLY_DECLINE
	} else if friendSettings.FriendVerifyType == apimodels.FriendVerifyType_NeedFriendVerify {
		storage := storages.NewFriendApplicationStorage()
		storage.Upsert(models.FriendApplication{
			RecipientId: req.FriendId,
			SponsorId:   userId,
			ApplyTime:   time.Now().UnixMilli(),
			Status:      models.FriendApplicationStatus(models.FriendApplicationStatus_Apply),
			AppKey:      appkey,
		})
		//send notify msg
		SendFriendApplyNotify(ctx, req.FriendId, &apimodels.FriendApplyNotify{
			SponsorId:   userId,
			RecipientId: req.FriendId,
		})
	} else if friendSettings.FriendVerifyType == apimodels.FriendVerifyType_NoNeedFriendVerify {
		fStorage.BatchUpsert([]models.FriendRel{
			{
				AppKey:   appkey,
				UserId:   userId,
				FriendId: req.FriendId,
			},
			{
				AppKey:   appkey,
				UserId:   req.FriendId,
				FriendId: userId,
			},
		})
		// sync to im
		if sdk := imsdk.GetImSdk(appkey); sdk != nil {
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    userId,
				FriendIds: []string{req.FriendId},
			})
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    req.FriendId,
				FriendIds: []string{userId},
			})
		}
		//send notify msg
		SendFriendNotify(ctx, req.FriendId, &apimodels.FriendNotify{
			Type: 0,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func ConfirmFriend(ctx context.Context, req *apimodels.ConfirmFriend) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendApplicationStorage()
	if req.IsAgree {
		fStorage := storages.NewFriendRelStorage()
		fStorage.BatchUpsert([]models.FriendRel{
			{
				AppKey:   appkey,
				UserId:   userId,
				FriendId: req.SponsorId,
			},
			{
				AppKey:   appkey,
				UserId:   req.SponsorId,
				FriendId: userId,
			},
		})
		// sync to im
		if sdk := imsdk.GetImSdk(appkey); sdk != nil {
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    userId,
				FriendIds: []string{req.SponsorId},
			})
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId:    req.SponsorId,
				FriendIds: []string{userId},
			})
		}
		//send notify msg
		SendFriendNotify(ctx, req.SponsorId, &apimodels.FriendNotify{
			Type: 1,
		})
		storage.UpdateStatus(appkey, req.SponsorId, userId, models.FriendApplicationStatus_Agree)
	} else {
		storage.UpdateStatus(appkey, req.SponsorId, userId, models.FriendApplicationStatus_Decline)
	}
	return errs.IMErrorCode_SUCCESS
}

func checkFriend(ctx context.Context, userId, friendId string) bool {
	results := CheckFriends(ctx, userId, []string{friendId})
	if isFriend, exist := results[friendId]; exist {
		return isFriend
	}
	return false
}

func checkBlockUser(ctx context.Context, userId, targetId string) bool {
	results := CheckBlockUsers(ctx, userId, []string{targetId})
	if isBLock, exist := results[targetId]; exist {
		return isBLock
	}
	return false
}

func QryMyFriendApplications(ctx context.Context, startTime int64, count int32, order int32) (errs.IMErrorCode, *apimodels.QryFriendApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendApplicationStorage()
	ret := &apimodels.QryFriendApplicationsResp{
		Items: []*apimodels.FriendApplicationItem{},
	}
	applications, err := storage.QueryMyApplications(appkey, userId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.FriendApplicationItem{
				Recipient: &apimodels.UserObj{
					UserId: application.RecipientId,
				},
				Status:    int32(application.Status),
				ApplyTime: application.ApplyTime,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryMyPendingFriendApplications(ctx context.Context, startTime int64, count int32, order int32) (errs.IMErrorCode, *apimodels.QryFriendApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendApplicationStorage()
	ret := &apimodels.QryFriendApplicationsResp{
		Items: []*apimodels.FriendApplicationItem{},
	}
	applications, err := storage.QueryPendingApplications(appkey, userId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			ret.Items = append(ret.Items, &apimodels.FriendApplicationItem{
				Sponsor: &apimodels.UserObj{
					UserId: application.SponsorId,
				},
				Status:    int32(application.Status),
				ApplyTime: application.ApplyTime,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func QryFriendApplications(ctx context.Context, startTime, count int64, order int32) (errs.IMErrorCode, *apimodels.QryFriendApplicationsResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendApplicationStorage()
	ret := &apimodels.QryFriendApplicationsResp{
		Items: []*apimodels.FriendApplicationItem{},
	}
	applications, err := storage.QueryApplications(appkey, userId, startTime, int64(count), order > 0)
	if err == nil {
		for _, application := range applications {
			item := &apimodels.FriendApplicationItem{
				Status:    int32(application.Status),
				ApplyTime: application.ApplyTime,
			}
			if userId == application.SponsorId {
				item.IsSponsor = true
				item.TargetUser = GetUser(ctx, application.RecipientId)
			} else {
				item.TargetUser = GetUser(ctx, application.SponsorId)
			}
			ret.Items = append(ret.Items, item)
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func CheckFriends(ctx context.Context, userId string, friendIds []string) map[string]bool {
	ret := make(map[string]bool)
	if len(friendIds) <= 0 {
		return ret
	}
	for _, friend := range friendIds {
		ret[friend] = false
	}
	storage := storages.NewFriendRelStorage()
	rels, err := storage.QueryFriendRelsByFriendIds(ctxs.GetAppKeyFromCtx(ctx), userId, friendIds)
	if err == nil {
		for _, rel := range rels {
			ret[rel.FriendId] = true
		}
	}
	return ret
}

func CheckBlockUsers(ctx context.Context, userId string, targetIds []string) map[string]bool {
	ret := make(map[string]bool)
	if len(targetIds) <= 0 {
		return ret
	}
	for _, uId := range targetIds {
		ret[uId] = false
	}
	storage := storages.NewBlockUserStorage()
	blockUsers, err := storage.FindBlockUserByIds(ctxs.GetAppKeyFromCtx(ctx), userId, targetIds)
	if err == nil {
		for _, item := range blockUsers {
			ret[item.BlockUserId] = true
		}
	}
	return ret
}

func SetFriendDisplayName(ctx context.Context, req *apimodels.SetFriendDisplayNameReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewFriendRelStorage()
	err := storage.UpdateDisplayName(appkey, userId, req.FriendId, req.FriendDisplayName)
	if err == nil {
		sdk := imsdk.GetImSdk(appkey)
		if sdk != nil {
			sdk.AddFriends(juggleimsdk.FriendIds{
				UserId: userId,
				Friends: []*juggleimsdk.FriendItem{
					{
						UserId:      userId,
						FriendId:    req.FriendId,
						DisplayName: req.FriendDisplayName,
					},
				},
			})
		}
	}
	return errs.IMErrorCode_SUCCESS
}
