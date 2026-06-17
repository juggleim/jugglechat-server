package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/emailengines"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/imsdk"
	"github.com/juggleim/jugglechat-server/commons/smsengines"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/events"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
)

var kefuInitFlags *sync.Map

func init() {
	kefuInitFlags = &sync.Map{}
	events.RegisteUserRegisteEvent(Welcome)
	events.RegisteUserRegisteEvent(JoinFeedbackGroup)
}

func Welcome(appkey string, user models.User) {
	if appkey != "" {
		sdk := imsdk.GetImSdk(appkey)
		if sdk != nil {
			appinfo, exist := appinfos.GetAppInfo(appkey)
			if exist && appinfo != nil {
				if _, loaded := kefuInitFlags.LoadOrStore(appkey, true); !loaded {
					sdk.Register(juggleimsdk.User{
						UserId:   "kefu",
						Nickname: "官方客服",
					})
				}
				if exist, val := appinfo.GetExt("welcome_msg"); exist {
					sdk.SendPrivateMsg(juggleimsdk.Message{
						SenderId:   "kefu",
						TargetIds:  []string{user.UserId},
						MsgType:    "jg:text",
						MsgContent: fmt.Sprintf(`{"content":"%s"}`, val),
					})
				}
			}
		}
	}
}

func JoinFeedbackGroup(appkey string, user models.User) {
	if appkey != "" {
		appInfo, exist := appinfos.GetAppInfo(appkey)
		if exist && appInfo != nil {
			exist, val := appInfo.GetExt("open_feedback_group")
			if exist && val != nil {
				valStr, ok := val.(string)
				if ok {
					b, err := strconv.ParseBool(valStr)
					if err == nil && b {
						feedbackGrpId := "feedbackgrp"
						grpStorage := storages.NewGroupStorage()
						grpInfo, err := grpStorage.FindById(appkey, feedbackGrpId)
						ctx := context.Background()
						ctx = context.WithValue(ctx, ctxs.CtxKey_AppKey, appkey)
						ctx = context.WithValue(ctx, ctxs.CtxKey_RequesterId, "kefu")
						if err == nil && grpInfo != nil { //grp existed
							AddGrpMembers(ctx, &apimodels.GroupMembersReq{
								GroupId:       feedbackGrpId,
								GroupName:     grpInfo.GroupName,
								GroupPortrait: grpInfo.GroupPortrait,
								MemberIds:     []string{user.UserId},
							})
						} else { //grp not existed
							CreateGroup(ctx, &apimodels.GroupMembersReq{
								GroupId:   feedbackGrpId,
								GroupName: "意见反馈群",
								MemberIds: []string{user.UserId},
							})
						}
					}
				}
			}
		}
	}
}

func QryUserInfo(ctx context.Context, userId string) (errs.IMErrorCode, *apimodels.UserObj) {
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	ret := &apimodels.UserObj{
		UserId: userId,
	}
	user := GetUser(ctx, userId)
	if user != nil {
		ret.Nickname = user.Nickname
		ret.Avatar = user.Avatar
		ret.UserType = user.UserType
		ret.Pinyin = user.Pinyin
		ret.Account = user.Account
		ret.Phone = user.Phone
		ret.Email = user.Email
	}
	if userId == requestId {
		ret.Settings = GetUserSettings(ctx, userId)
	} else {
		ret.IsFriend = checkFriend(ctx, requestId, userId)
		ret.IsBlock = checkBlockUser(ctx, requestId, userId)
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func GetUserSettings(ctx context.Context, userId string) *apimodels.UserSettings {
	settings := &apimodels.UserSettings{}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewUserExtStorage()
	exts, err := storage.QryExtFields(appkey, userId)
	if err == nil {
		settings.FriendVerifyType = apimodels.FriendVerifyType_NeedFriendVerify
		for _, ext := range exts {
			if ext.ItemKey == apimodels.UserExtKey_Language {
				settings.Language = ext.ItemValue
			} else if ext.ItemKey == apimodels.UserExtKey_Undisturb {
				settings.Undisturb = ext.ItemValue
			} else if ext.ItemKey == apimodels.UserExtKey_FriendVerifyType {
				verifyType := utils.ToInt(ext.ItemValue)
				settings.FriendVerifyType = verifyType
			} else if ext.ItemKey == apimodels.UserExtKey_GrpVerifyType {
				verifyType := utils.ToInt(ext.ItemValue)
				settings.GrpVerifyType = verifyType
			}
		}
	}
	return settings
}

func SearchByPhone(ctx context.Context, phone string) (errs.IMErrorCode, *apimodels.Users) {
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	users := &apimodels.Users{
		Items: []*apimodels.UserObj{},
	}
	targetUserId := utils.ShortMd5(phone)
	storage := storages.NewUserStorage()
	user, err := storage.FindByPhone(appkey, phone)
	if err == nil && user != nil {
		targetUserId = user.UserId
		users.Items = append(users.Items, &apimodels.UserObj{
			UserId:   user.UserId,
			Nickname: user.Nickname,
			Avatar:   user.UserPortrait,
			UserType: user.UserType,
			IsFriend: checkFriend(ctx, requestId, targetUserId),
		})
	} else {
		user, err := storage.FindByUserId(appkey, targetUserId)
		if err == nil && user != nil {
			users.Items = append(users.Items, &apimodels.UserObj{
				UserId:   user.UserId,
				Nickname: user.Nickname,
				Avatar:   user.UserPortrait,
				UserType: user.UserType,
				IsFriend: checkFriend(ctx, requestId, targetUserId),
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, users
}

func SearchByKeyword(ctx context.Context, keyword string) (errs.IMErrorCode, *apimodels.Users) {
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	ret := &apimodels.Users{
		Items: []*apimodels.UserObj{},
	}
	storage := storages.NewUserStorage()
	users, err := storage.SearchByKeyword(appkey, requestId, keyword, false)
	if err == nil {
		targetUIds := []string{}
		for _, user := range users {
			targetUIds = append(targetUIds, user.UserId)
			ret.Items = append(ret.Items, &apimodels.UserObj{
				UserId:   user.UserId,
				Nickname: user.Nickname,
				Avatar:   user.UserPortrait,
				UserType: user.UserType,
			})
		}
		isFriendMap := CheckFriends(ctx, requestId, targetUIds)
		for _, user := range ret.Items {
			user.IsFriend = isFriendMap[user.UserId]
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SetUserAccount(ctx context.Context, req *apimodels.SetUserAccountReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewUserStorage()
	err := storage.UpdateAccount(appkey, requestId, req.Account)
	if err != nil {
		return errs.IMErrorCode_APP_USER_EXISTED
	}
	return errs.IMErrorCode_SUCCESS
}

func UpdateUser(ctx context.Context, req *apimodels.UserObj) errs.IMErrorCode {
	if ok, _ := CheckSensitiveText(ctx, req.Nickname); !ok {
		return errs.IMErrorCode_APP_Sensitive
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewUserStorage()
	storage.Update(appkey, req.UserId, req.Nickname, req.Avatar)
	// sync to imserver
	sdk := imsdk.GetImSdk(appkey)
	if sdk != nil {
		sdk.Register(juggleimsdk.User{
			UserId:       req.UserId,
			Nickname:     req.Nickname,
			UserPortrait: req.Avatar,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func UpdatePass(ctx context.Context, req *apimodels.UpdUserPassReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewUserStorage()
	user, err := storage.FindByUserId(appkey, req.UserId)
	if err != nil || user == nil {
		return errs.IMErrorCode_APP_USER_NOT_EXIST
	}
	if user.LoginPass != utils.SHA1(req.Password) {
		return errs.IMErrorCode_APP_LOGIN_ERR_PASS
	}
	err = storage.UpdatePass(appkey, req.UserId, utils.SHA1(req.NewPassword))
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}

func UpdateUserSettings(ctx context.Context, req *apimodels.UserSettings) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requestId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewUserExtStorage()
	settings := map[juggleimsdk.UserSettingKey]string{}
	if req.Language != "" {
		storage.Upsert(models.UserExt{
			UserId:    requestId,
			ItemKey:   apimodels.UserExtKey_Language,
			ItemValue: req.Language,
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
		settings[juggleimsdk.UserSettingKey_Language] = req.Language
	}
	if req.Undisturb != "" {
		storage.Upsert(models.UserExt{
			UserId:    requestId,
			ItemKey:   apimodels.UserExtKey_Undisturb,
			ItemValue: req.Undisturb,
			ItemType:  apimodels.AttItemType_Setting,
			AppKey:    appkey,
		})
		settings[juggleimsdk.UserSettingKey_Undisturb] = req.Undisturb
	}
	storage.Upsert(models.UserExt{
		UserId:    requestId,
		ItemKey:   apimodels.UserExtKey_FriendVerifyType,
		ItemValue: utils.Int2String(int64(req.FriendVerifyType)),
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	storage.Upsert(models.UserExt{
		UserId:    requestId,
		ItemKey:   apimodels.UserExtKey_GrpVerifyType,
		ItemValue: utils.Int2String(int64(req.GrpVerifyType)),
		ItemType:  apimodels.AttItemType_Setting,
		AppKey:    appkey,
	})
	//sync to im
	if len(settings) > 0 {
		if sdk := imsdk.GetImSdk(appkey); sdk != nil {
			sdk.SetUserSettings(juggleimsdk.User{
				UserId:   requestId,
				Settings: settings,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS
}

func QueryMyGroups(ctx context.Context, limit int64, offset string) (errs.IMErrorCode, *apimodels.Groups) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	memberId := ctxs.GetRequesterIdFromCtx(ctx)
	dao := dbs.GroupMemberDao{}
	var startId int64
	if offset != "" {
		startId, _ = utils.DecodeInt(offset)
	}
	groups, err := dao.QueryGroupsByMemberId(appkey, memberId, startId, limit)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT, nil
	}
	ret := &apimodels.Groups{
		Items: []*apimodels.Group{},
	}
	for _, group := range groups {
		ret.Offset, err = utils.EncodeInt(group.ID)
		if err == nil {
			ret.Items = append(ret.Items, &apimodels.Group{
				GroupId:       group.GroupId,
				GroupName:     group.GroupName,
				GroupPortrait: group.GroupPortrait,
			})
		} else {
			fmt.Println(err)
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func SearchMyGroups(ctx context.Context, req *apimodels.SearchReq) (errs.IMErrorCode, *apimodels.Groups) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	memberId := ctxs.GetRequesterIdFromCtx(ctx)
	dao := dbs.GroupMemberDao{}
	var startId int64
	if req.Offset != "" {
		startId, _ = utils.DecodeInt(req.Offset)
	}
	groups, err := dao.SearchGroupsByMemberId(appkey, memberId, req.Keyword, startId, req.Limit)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT, nil
	}
	ret := &apimodels.Groups{
		Items: []*apimodels.Group{},
	}
	for _, group := range groups {
		ret.Offset, err = utils.EncodeInt(group.ID)
		if err == nil {
			ret.Items = append(ret.Items, &apimodels.Group{
				GroupId:       group.GroupId,
				GroupName:     group.GroupName,
				GroupPortrait: group.GroupPortrait,
			})
		} else {
			fmt.Println(err)
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func GetUser(ctx context.Context, userId string) *apimodels.UserObj {
	if userId == "" {
		return nil
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	u := &apimodels.UserObj{
		UserId: userId,
	}
	storage := storages.NewUserStorage()
	user, err := storage.FindByUserId(appkey, userId)
	if err == nil && user != nil {
		u.Nickname = user.Nickname
		u.Avatar = user.UserPortrait
		u.UserType = user.UserType
		u.Pinyin = user.Pinyin
		u.Account = user.LoginAccount
		u.Phone = utils.MaskPhone(user.Phone)
		u.Email = utils.MaskEmail(user.Email)
	}
	return u
}

func BlockUsers(ctx context.Context, req *apimodels.BlockUsersReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewBlockUserStorage()
	for _, blockId := range req.BlockUserIds {
		storage.Create(models.BlockUser{
			UserId:      userId,
			BlockUserId: blockId,
			AppKey:      appkey,
		})
	}
	return errs.IMErrorCode_SUCCESS
}

func UnBlockUsers(ctx context.Context, req *apimodels.BlockUsersReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewBlockUserStorage()
	storage.BatchDelBlockUsers(appkey, userId, req.BlockUserIds)
	return errs.IMErrorCode_SUCCESS
}

func QryBlockUsers(ctx context.Context, limit int64, offset string) (errs.IMErrorCode, *apimodels.BlockUsers) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	storage := storages.NewBlockUserStorage()
	ret := &apimodels.BlockUsers{
		Items: []*apimodels.UserObj{},
	}
	var startId int64 = 0
	if offset != "" {
		intVal, err := utils.DecodeInt(offset)
		if err == nil && intVal > 0 {
			startId = intVal
		}
	}
	users, err := storage.QryBlockUsers(appkey, userId, limit, startId)
	if err == nil {
		for _, user := range users {
			o, err := utils.EncodeInt(user.ID)
			if err == nil && o != "" {
				ret.Offset = o
			}
			ret.Items = append(ret.Items, &apimodels.UserObj{
				UserId:   user.BlockUserId,
				Pinyin:   user.Pinyin,
				Nickname: user.Nickname,
				Avatar:   user.UserPortrait,
				UserType: user.UserType,
				IsBlock:  true,
			})
		}
	}
	return errs.IMErrorCode_SUCCESS, ret
}

func BindEmailSendEmail(ctx context.Context, req *apimodels.BindEmailReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewUserStorage()
	user, err := storage.FindByEmail(appkey, req.Email)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	if user != nil {
		return errs.IMErrorCode_APP_EMAIL_EXIST
	}
	mailEngin := GetMailEngine(appkey)
	if mailEngin == nil || mailEngin == emailengines.DefaultEmailEngine {
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	// 检查是否还有有效的
	recordStorage := storages.NewSmsRecordStorage()
	record, err := recordStorage.FindByEmail(appkey, req.Email, time.Now().Add(-3*time.Minute))
	randomCode := RandomSms()
	if err == nil {
		randomCode = record.Code
	} else {
		_, err = recordStorage.Create(models.SmsRecord{
			AppKey:      appkey,
			Email:       req.Email,
			Code:        randomCode,
			CreatedTime: time.Now(),
		})
		if err != nil {
			return errs.IMErrorCode_APP_SMS_SEND_FAILED
		}
	}
	body, html := GetEmailTemplate()
	if html != "" {
		body = ""
		html = strings.ReplaceAll(html, "{code}", randomCode)
	} else {
		html = ""
		body = strings.ReplaceAll(body, "{code}", randomCode)
	}
	err = mailEngin.SendMail(req.Email, "Verify Code", body, html)
	if err != nil {
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	return errs.IMErrorCode_SUCCESS
}

func BindEmail(ctx context.Context, req *apimodels.BindEmailReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	if req.Code != "000000" {
		storage := storages.NewSmsRecordStorage()
		record, err := storage.FindByEmailCode(appkey, req.Email, req.Code)
		if err != nil {
			return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
		}
		interval := time.Since(record.CreatedTime)
		if interval > 5*time.Minute {
			return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
		}
	}
	//update email
	userStorage := storages.NewUserStorage()
	err := userStorage.UpdateEmail(appkey, userId, req.Email)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}

func BindPhoneSendSms(ctx context.Context, req *apimodels.BindPhoneReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewUserStorage()
	user, err := storage.FindByPhone(appkey, req.Phone)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	if user != nil {
		return errs.IMErrorCode_APP_PHONE_EXISTED
	}
	smsEngin := GetSmsEngine(appkey)
	if smsEngin == nil || smsEngin == smsengines.DefaultSmsEngine {
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	// 检查是否还有有效的
	recordStorage := storages.NewSmsRecordStorage()
	record, err := recordStorage.FindByPhone(appkey, req.Phone, time.Now().Add(-3*time.Minute))
	randomCode := RandomSms()
	if err == nil {
		randomCode = record.Code
	} else {
		_, err = recordStorage.Create(models.SmsRecord{
			AppKey:      appkey,
			Phone:       req.Phone,
			Code:        randomCode,
			CreatedTime: time.Now(),
		})
		if err != nil {
			return errs.IMErrorCode_APP_SMS_SEND_FAILED
		}
	}
	err = smsEngin.SmsSend(req.Phone, map[string]interface{}{
		"code": randomCode,
	})
	if err != nil {
		return errs.IMErrorCode_APP_SMS_SEND_FAILED
	}
	return errs.IMErrorCode_SUCCESS
}

func BindPhone(ctx context.Context, req *apimodels.BindPhoneReq) errs.IMErrorCode {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	userId := ctxs.GetRequesterIdFromCtx(ctx)
	if req.Code != "000000" {
		storage := storages.NewSmsRecordStorage()
		record, err := storage.FindByPhoneCode(appkey, req.Phone, req.Code)
		if err != nil {
			return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
		}
		interval := time.Since(record.CreatedTime)
		if interval > 5*time.Minute {
			return errs.IMErrorCode_APP_SMS_CODE_EXPIRED
		}
	}
	//update phone
	userStorage := storages.NewUserStorage()
	err := userStorage.UpdatePhone(appkey, userId, req.Phone)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT
	}
	return errs.IMErrorCode_SUCCESS
}
