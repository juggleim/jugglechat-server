package models

type CustomChatMsgReq struct {
	Sender      string       `json:"sender"`
	Receiver    string       `json:"receiver"`
	ConverType  int          `json:"conver_type"`
	MsgType     string       `json:"msg_type"`
	MsgContent  string       `json:"msg_content"`
	MsgId       string       `json:"msg_id"`
	MsgTime     int64        `json:"msg_time"`
	MentionInfo *MentionInfo `json:"mention_info,omitempty"`
}

type MentionInfo struct {
	MentionType   string   `json:"mention_type"`
	TargetUserIds []string `json:"target_user_ids"`
}
