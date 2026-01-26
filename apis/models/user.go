package models

var (
	UserExtKey_Phone            string = "phone"
	UserExtKey_Language         string = "language"
	UserExtKey_Undisturb        string = "undisturb"
	UserExtKey_FriendVerifyType string = "friend_verify_type"
	UserExtKey_GrpVerifyType    string = "grp_verify_type"
)

const (
	AttItemType_Att     int = 0
	AttItemType_Setting int = 1
	AttItemType_Status  int = 2
)

type UserObj struct {
	UserId     string        `json:"user_id"`
	Nickname   string        `json:"nickname"`
	Avatar     string        `json:"avatar"`
	Pinyin     string        `json:"pinyin"`
	UserType   int           `json:"user_type"`
	Phone      string        `json:"phone"`
	Email      string        `json:"email"`
	Account    string        `json:"account"`
	Status     int32         `json:"status"`
	IsFriend   bool          `json:"is_friend"`
	FriendInfo *FriendInfo   `json:"friend_info"`
	IsBlock    bool          `json:"is_block"`
	Settings   *UserSettings `json:"settings"`
}

type FriendInfo struct {
	IsFriend    bool   `json:"is_friend"`
	DisplayName string `json:"display_name"`
}

type UserSettings struct {
	Language         string `json:"language"`
	FriendVerifyType int    `json:"friend_verify_type"`
	GrpVerifyType    int    `json:"grp_verify_type"`
	Undisturb        string `json:"undisturb"`
}

type Users struct {
	Items  []*UserObj `json:"items"`
	Offset string     `json:"offset"`
}

type UserIds struct {
	UserIds []string `json:"user_ids"`
}

type SearchReq struct {
	Keyword string `json:"keyword"`
	Phone   string `json:"phone"`
	Limit   int64  `json:"limit"`
	Offset  string `json:"offset"`
}

type SetUserAccountReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type UpdUserPassReq struct {
	UserId      string `json:"user_id"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
}

type Friends struct {
	Items  []*UserObj `json:"items"`
	Offset string     `json:"offset,omitempty"`
}

type Friend struct {
	UserId   string `json:"user_id"`
	FriendId string `json:"friend_id"`
}

type FriendIds struct {
	FriendIds []string `json:"friend_ids"`
}

type UserConfs struct {
}

// user block
type BlockUsersReq struct {
	BlockUserIds []string `json:"block_user_ids"`
}

type BlockUsers struct {
	Items  []*UserObj `json:"items"`
	Offset string     `json:"offset,omitempty"`
}

type BindEmailReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type BindPhoneReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
