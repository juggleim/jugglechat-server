package models

const (
	FriendVerifyType_NoNeedFriendVerify int = 0
	FriendVerifyType_NeedFriendVerify   int = 1
	FriendVerifyType_DeclineFriend      int = 2
)

type FriendIdsReq struct {
	FriendIds []string `json:"friend_ids"`
}

type FriendMember struct {
	FriendId string `json:"friend_id"`
	OrderTag string `json:"order_tag"`
}

type ApplyFriend struct {
	FriendId string `json:"friend_id"`
}

type ConfirmFriend struct {
	SponsorId string `json:"sponsor_id"`
	IsAgree   bool   `json:"is_agree"`
}

type QryFriendApplicationsResp struct {
	Items []*FriendApplicationItem `json:"items"`
}

type FriendApplicationItem struct {
	Recipient  *UserObj `json:"recipient,omitempty"`
	Sponsor    *UserObj `json:"sponsor,omitempty"`
	TargetUser *UserObj `json:"target_user,omitempty"`
	IsSponsor  bool     `json:"is_sponsor,omitempty"`
	Status     int32    `json:"status,omitempty"`
	ApplyTime  int64    `json:"apply_time,omitempty"`
}

type SearchFriendsReq struct {
	Key    string `json:"key"`
	Offset string `json:"offset"`
	Limit  int64  `json:"limit"`
}

type SetFriendDisplayNameReq struct {
	FriendId          string `json:"friend_id"`
	FriendDisplayName string `json:"friend_display_name"`
}
