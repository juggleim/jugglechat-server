package services

import (
	"appserver/dbs"
	"appserver/utils"

	imsdk "github.com/juggleim/imserver-sdk-go"
)

func AddFriend(userId, friendId string) ErrorCode {
	var userIdInt, friendIdInt int64
	userIdInt, err := utils.Decode(userId)
	if err != nil || userIdInt == 0 {
		return ErrorCode_IdDecodeFail
	}
	friendIdInt, err = utils.Decode(friendId)
	if err != nil {
		return ErrorCode_IdDecodeFail
	}
	friendDao := dbs.FriendDao{}
	friends := []dbs.FriendDao{}
	friends = append(friends, dbs.FriendDao{
		UserId:   userIdInt,
		FriendId: friendIdInt,
	})
	friends = append(friends, dbs.FriendDao{
		UserId:   friendIdInt,
		FriendId: userIdInt,
	})
	err = friendDao.BatchCreate(friends)
	if err != nil {
		return ErrorCode_UserDbUpdateFail
	}
	//send notify msg
	notify := &FriendNotify{
		Type: 0,
	}
	SendPrivateMsg(imsdk.Message{
		SenderId:       userId,
		TargetIds:      []string{friendId},
		MsgType:        FriendNotifyMsgType,
		MsgContent:     utils.ToJson(notify),
		IsStorage:      utils.BoolPtr(true),
		IsCount:        utils.BoolPtr(false),
		IsNotifySender: utils.BoolPtr(true),
	})
	return ErrorCode_Success

}

var FriendNotifyMsgType string = "jgd:friendntf"

type FriendNotify struct {
	Type int `json:"type"`
}

func QryFrineds(userId, startId string, count int) (ErrorCode, *Friends) {
	friendDao := dbs.FriendDao{}
	userIdInt, err := utils.Decode(userId)
	if err != nil {
		return ErrorCode_IdDecodeFail, nil
	}
	var startIdInt int64 = 0
	if startId != "" {
		startIdInt, err = utils.Decode(startId)
		if err != nil {
			startIdInt = 0
		}
	}
	friends, err := friendDao.QueryFriends(userIdInt, startIdInt, int64(count))
	if err != nil {
		return ErrorCode_UserDbReadFail, nil
	}
	resp := &Friends{
		Items: []*User{},
	}
	userDao := dbs.UserDao{}
	for _, friend := range friends {
		userdb := userDao.FindByUserId(friend.FriendId)
		if userdb != nil {
			idStr, _ := utils.Encode(userdb.ID)
			fri := &User{
				UserId:   idStr,
				Nickname: userdb.Nickname,
				Avatar:   userdb.Avatar,
				Status:   userdb.Status,
				Phone:    userdb.Phone,
			}
			resp.Items = append(resp.Items, fri)
		}
	}
	return ErrorCode_Success, resp
}

type Friends struct {
	Items []*User `json:"items"`
}

type Friend struct {
	UserId   string `json:"user_id"`
	FriendId string `json:"friend_id"`
}
