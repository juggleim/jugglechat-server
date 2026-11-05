package models

import "github.com/juggleim/jugglechat-server/services/deviceauth"

type RegisterReq struct {
	deviceauth.DeviceParams
	Account  string `json:"account"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type SmsLoginReq struct {
	deviceauth.DeviceParams
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type EmailLoginReq struct {
	deviceauth.DeviceParams
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginUserResp struct {
	UserId        string `json:"user_id"`
	Authorization string `json:"authorization"`
	NickName      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Status        int    `json:"status"`
	ImToken       string `json:"im_token,omitempty"`
}

type QrCode struct {
	deviceauth.DeviceParams
	PollSecret string `json:"poll_secret"`
	Id         string `json:"id"`
}
