package services

import (
	"context"
	"time"

	"github.com/juggleim/commons/errs"
	"github.com/juggleim/commons/tools"
	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
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
	storage := storages.NewBanUserStorage()
	appkey := req.AppKey
	for _, user := range req.Items {
		var endTime int64 = user.EndTime
		if endTime == 0 && user.EndTimeOffset > 0 {
			endTime = time.Now().UnixMilli() + user.EndTimeOffset
		}
		storage.Upsert(models.BanUser{
			UserId:      user.UserId,
			CreatedTime: time.Now(),
			EndTime:     endTime,
			AppKey:      appkey,
		})
	}
	return errs.AdminErrorCode_Success
}

func UnBanUsers(ctx context.Context, req *apimodels.BanUsersReq) errs.AdminErrorCode {
	storage := storages.NewBanUserStorage()
	appkey := req.AppKey
	for _, user := range req.Items {
		storage.DelBanUser(appkey, user.UserId, "")
	}
	return errs.AdminErrorCode_Success
}
