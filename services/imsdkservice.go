package services

import (
	"appserver/configures"
	"appserver/serversdk"
	"appserver/utils"
	"errors"
	"fmt"
)

var imsdk *serversdk.JuggleIMSdk

func InitImSdk() error {
	imsdk = serversdk.NewJuggleIMSdk(configures.Config.Im.AppKey, configures.Config.Im.AppSecret, configures.Config.Im.ApiUrl)
	if imsdk != nil {
		return nil
	} else {
		return errors.New("init im sdk failed")
	}
}

func RegisterImToken(u serversdk.User) string {
	resp, code, _, err := imsdk.Register(u)
	if err == nil && code == serversdk.ApiCode_Success {
		return resp.Token
	}
	return ""
}

func CreateGroup2Im(req serversdk.GroupMembersReq) ErrorCode {
	code, _, err := imsdk.CreateGroup(req)
	if err == nil && code == serversdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func AddGroupMembers2Im(req serversdk.GroupMembersReq) ErrorCode {
	code, _, err := imsdk.GroupAddMembers(req)
	if err == nil && code == serversdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}
func DelGroupMembers2Im(req serversdk.GroupMembersReq) ErrorCode {
	code, _, err := imsdk.GroupDelMembers(req)
	if err == nil && code == serversdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func UpdateGroupInfo2Im(groupInfo serversdk.GroupInfo) ErrorCode {
	code, _, err := imsdk.UpdateGroup(groupInfo)
	if err == nil && code == serversdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func SendGroupMsg(msg serversdk.ImMessage) ErrorCode {
	code, _, err := imsdk.SendGroupMsg(msg)
	fmt.Println("send grp msg:", code, err, utils.ToJson(msg))
	if err == nil && code == serversdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}
