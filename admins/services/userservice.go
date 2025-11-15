package services

import (
	"context"
	"time"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"
)

func QryUsers(ctx context.Context, appkey, userId, name, offset string, limit int64, isPositive bool) (errs.AdminErrorCode, *apimodels.Users) {
	var startId int64 = 0
	var err error
	if offset != "" {
		startId, err = tools.DecodeInt(offset)
		if err != nil {
			startId = 0
		}
	}
	ret := &apimodels.Users{
		Items: []*apimodels.User{},
	}
	storage := storages.NewUserStorage()
	if userId != "" {
		user, err := storage.FindByUserId(appkey, userId)
		if err == nil && user != nil {
			ret.Items = append(ret.Items, &apimodels.User{
				UserId:      user.UserId,
				Nickname:    user.Nickname,
				Avatar:      user.UserPortrait,
				Pinyin:      user.Pinyin,
				UserType:    user.UserType,
				Status:      int32(user.Status),
				CreatedTime: user.CreatedTime.UnixMilli(),
			})
		}
	} else {
		users, err := storage.QryUsers(appkey, name, startId, limit, isPositive)
		if err == nil {
			for _, user := range users {
				ret.Offset, _ = tools.EncodeInt(user.ID)
				ret.Items = append(ret.Items, &apimodels.User{
					UserId:      user.UserId,
					Nickname:    user.Nickname,
					Avatar:      user.UserPortrait,
					Pinyin:      user.Pinyin,
					UserType:    user.UserType,
					Status:      int32(user.Status),
					CreatedTime: user.CreatedTime.UnixMilli(),
				})
			}
		}
	}
	return errs.AdminErrorCode_Success, ret
}

func QryUserInfo(appkey, userId string) *apimodels.User {
	storage := storages.NewUserStorage()
	user, err := storage.FindByUserId(appkey, userId)
	if err != nil || user == nil {
		return &apimodels.User{
			UserId: userId,
		}
	}
	return &apimodels.User{
		UserId:   user.UserId,
		Nickname: user.Nickname,
		UserType: user.UserType,
		Avatar:   user.UserPortrait,
	}
}

func BanUsers(ctx context.Context, req *apimodels.BanUsersReq) errs.AdminErrorCode {
	userStorage := storages.NewUserStorage()
	appkey := req.AppKey
	banUsers := &juggleimsdk.BanUsers{
		Items: []*juggleimsdk.BanUser{},
	}
	for _, user := range req.Items {
		var endTime int64 = user.EndTime
		if endTime == 0 && user.EndTimeOffset > 0 {
			endTime = time.Now().UnixMilli() + user.EndTimeOffset
		}
		banUsers.Items = append(banUsers.Items, &juggleimsdk.BanUser{
			UserId:  user.UserId,
			EndTime: endTime,
		})
		userStorage.UpdateStatus(appkey, user.UserId, models.UserStatus_Ban)
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.BanUsers(banUsers)
	}
	return errs.AdminErrorCode_Success
}

func UnBanUsers(ctx context.Context, req *apimodels.BanUsersReq) errs.AdminErrorCode {
	userStorage := storages.NewUserStorage()
	banUsers := &juggleimsdk.BanUsers{
		Items: []*juggleimsdk.BanUser{},
	}
	appkey := req.AppKey
	for _, user := range req.Items {
		banUsers.Items = append(banUsers.Items, &juggleimsdk.BanUser{
			UserId: user.UserId,
		})
		userStorage.UpdateStatus(appkey, user.UserId, models.UserStatus_Normal)
	}
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.UnBanUsers(banUsers)
	}
	return errs.AdminErrorCode_Success
}
