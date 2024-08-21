package services

import (
	"appserver/configures"
	"appserver/utils"
	"errors"
	"fmt"

	imsdk "github.com/juggleim/imserver-sdk-go"
)

var sdk *imsdk.JuggleIMSdk

func InitImSdk() error {
	sdk = imsdk.NewJuggleIMSdk(configures.Config.Im.AppKey, configures.Config.Im.AppSecret, configures.Config.Im.ApiUrl)
	if sdk != nil {
		return nil
	} else {
		return errors.New("init im sdk failed")
	}
}

func RegisterImToken(u imsdk.User) string {
	resp, code, _, err := sdk.Register(u)
	if err == nil && code == imsdk.ApiCode_Success {
		return resp.Token
	}
	return ""
}

func CreateGroup2Im(req imsdk.GroupMembersReq) ErrorCode {
	code, _, err := sdk.CreateGroup(req)
	if err == nil && code == imsdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func AddGroupMembers2Im(req imsdk.GroupMembersReq) ErrorCode {
	code, _, err := sdk.GroupAddMembers(req)
	if err == nil && code == imsdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}
func DelGroupMembers2Im(req imsdk.GroupMembersReq) ErrorCode {
	code, _, err := sdk.GroupDelMembers(req)
	if err == nil && code == imsdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func UpdateGroupInfo2Im(groupInfo imsdk.GroupInfo) ErrorCode {
	code, _, err := sdk.UpdateGroup(groupInfo)
	if err == nil && code == imsdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}

func SendGroupMsg(msg imsdk.Message) ErrorCode {
	code, _, err := sdk.SendGroupMsg(msg)
	fmt.Println("send grp msg:", code, err, utils.ToJson(msg))
	if err == nil && code == imsdk.ApiCode_Success {
		return ErrorCode_Success
	}
	return ErrorCode_Sync2ImFail
}
