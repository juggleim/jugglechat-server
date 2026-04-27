package models

import juggleimsdk "github.com/juggleim/imserver-sdk-go"

type HisMsgs struct {
	Msgs []*HisMsg `json:"items"`
}

type HisMsg struct {
	Sender     *User  `json:"sender"`
	MsgId      string `json:"msg_id"`
	MsgTime    int64  `json:"msg_time"`
	MsgType    string `json:"msg_type"`
	MsgContent string `json:"msg_content"`
}

type RecallHisMsgReq struct {
	AppKey      string            `json:"app_key"`
	FromId      string            `json:"from_id"`
	TargetId    string            `json:"target_id"`
	ChannelType int               `json:"channel_type"`
	MsgId       string            `json:"msg_id"`
	MsgTime     int64             `json:"msg_time"`
	Exts        map[string]string `json:"exts"`
}

type DelHisMsgsReq struct {
	AppKey      string                   `json:"app_key"`
	FromId      string                   `json:"from_id"`
	TargetId    string                   `json:"target_id"`
	ChannelType int                      `json:"channel_type"`
	Msgs        []*juggleimsdk.SimpleMsg `json:"msgs"`
}
