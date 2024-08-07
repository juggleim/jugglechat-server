package serversdk

import "net/http"

type ImMessage struct {
	SenderId       string `json:"sender_id"`
	TargetId       string `json:"target_id"`
	MsgType        string `json:"msg_type"`
	MsgContent     string `json:"msg_content"`
	IsStorage      bool   `json:"is_storage"`
	IsCount        bool   `json:"is_count"`
	IsNotifySender bool   `json:"is_notify_sender"`
}

func (sdk *JuggleIMSdk) SendGroupMsg(msg ImMessage) (ApiCode, string, error) {
	url := sdk.ApiUrl + "/apigateway/messages/group/send"
	code, traceId, err := sdk.HttpCall(http.MethodPost, url, msg, nil)
	return code, traceId, err
}
