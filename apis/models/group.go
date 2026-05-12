package models

type Group struct {
	GroupId         string             `json:"group_id"`
	GroupName       string             `json:"group_name"`
	GroupPortrait   string             `json:"group_portrait"`
	GrpMembers      []*GroupMemberInfo `json:"members,omitempty"`
	MemberIds       []string           `json:"member_ids,omitempty"`
	MemberCount     int                `json:"member_count"`
	Owner           *UserObj           `json:"owner,omitempty"`
	MyRole          int                `json:"my_role"`
	GroupManagement *GroupManagement   `json:"group_management"`
}

type GroupManagement struct {
	GroupId            string `json:"group_id,omitempty"`
	GroupMute          int    `json:"group_mute"`
	MaxAdminCount      int    `json:"max_admin_count"`
	AdminCount         int    `json:"admin_count"`
	GroupVerifyType    int    `json:"group_verify_type"`
	GroupHisMsgVisible int    `json:"group_his_msg_visible"`

	GroupEditMsgRight     *int `json:"group_edit_msg_right"`
	GroupAddMemberRight   *int `json:"group_add_member_right"`
	GroupMentionAllRight  *int `json:"group_mention_all_right"`
	GroupTopMsgRight      *int `json:"group_top_msg_right"`
	GroupSendMsgRight     *int `json:"group_send_msg_right"`
	GroupSetMsgLifeRight  *int `json:"group_set_msg_life_right"`
	GroupApplyFriendRight *int `json:"group_apply_friend_right"`
}

type Groups struct {
	Items  []*Group `json:"items"`
	Offset string   `json:"offset,omitempty"`
}

type GroupAnnouncement struct {
	GroupId string `json:"group_id"`
	Content string `json:"content"`
}

type CheckGroupMembersReq struct {
	GroupId   string   `json:"group_id"`
	MemberIds []string `json:"member_ids"`
}

type CheckGroupMembersResp struct {
	GroupId        string          `json:"group_id"`
	MemberExistMap map[string]bool `json:"member_exist_map"`
}

type SearchGroupMembersReq struct {
	GroupId string `json:"group_id"`
	Key     string `json:"key"`
	Offset  string `json:"offset"`
	Limit   int64  `json:"limit"`
}

type KvItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	UpdTime int64  `json:"upd_time"`
}

type GroupMembersReq struct {
	GroupId       string    `json:"group_id"`
	GroupName     string    `json:"group_name"`
	GroupPortrait string    `json:"group_portrait"`
	MemberIds     []string  `json:"member_ids"`
	ExtFields     []*KvItem `json:"ext_fields"`
	Settings      []*KvItem `json:"settings"`
}

type GroupInfo struct {
	GroupId       string    `json:"group_id"`
	GroupName     string    `json:"group_name"`
	GroupPortrait string    `json:"group_portrait"`
	IsMute        int32     `json:"is_mute"`
	ExtFields     []*KvItem `json:"ext_fields"`
	UpdatedTime   int64     `json:"updated_time"`
	Settings      []*KvItem `json:"settings"`
	MemberCount   int32     `json:"member_count"`
}

type GroupInviteReq struct {
	GroupId   string   `json:"group_id"`
	MemberIds []string `json:"member_ids"`
}

type GrpInviteResultReason int32

const (
	GrpInviteResultReason_InviteSucc     GrpInviteResultReason = 0
	GrpInviteResultReason_InviteSendOut  GrpInviteResultReason = 1
	GrpInviteResultReason_InviteDecline  GrpInviteResultReason = 2
	GrpInviteResultReason_InviteRepeated GrpInviteResultReason = 3
)

type GroupInviteResp struct {
	Reason  GrpInviteResultReason            `json:"reason"`
	Results map[string]GrpInviteResultReason `json:"results"`
}

type GroupConfirm struct {
	ApplicationId string `json:"application_id"`
	IsAgree       bool   `json:"is_agree"`
}

const (
	GrpVerifyType_NoNeedGrpVerify int = 0
	GrpVerifyType_NeedGrpVerify   int = 1
	GrpVerifyType_DeclineGroup    int = 2
)

type GroupMemberInfos struct {
	Items  []*GroupMemberInfo `json:"items"`
	Offset string             `json:"offset"`
}

type GrpMemberRole int32

const (
	GrpMemberRole_GrpMember    GrpMemberRole = 0
	GrpMemberRole_GrpCreator   GrpMemberRole = 1
	GrpMemberRole_GrpAdmin     GrpMemberRole = 2
	GrpMemberRole_GrpNotMember GrpMemberRole = 3
)

type GroupMemberInfo struct {
	UserId         string        `json:"user_id"`
	Nickname       string        `json:"nickname"`
	Avatar         string        `json:"avatar"`
	GrpDisplayName string        `json:"grp_display_name,omitempty"`
	MemberType     int           `json:"member_type"`
	Role           GrpMemberRole `json:"role"`
	IsMute         int           `json:"is_mute"`
	FriendInfo     *FriendInfo   `json:"friend_info"`
}

type GroupOwnerChgReq struct {
	GroupId string `json:"group_id"`
	OwnerId string `json:"owner_id"`
}

type SetGroupMuteReq struct {
	GroupId string `json:"group_id"`
	IsMute  int32  `json:"is_mute"`
}

type SetGroupMemberMuteReq struct {
	GroupId   string   `json:"group_id"`
	MemberIds []string `json:"member_ids"`
	IsMute    int32    `json:"is_mute"`
}

type SetGroupVerifyTypeReq struct {
	GroupId    string `json:"group_id"`
	VerifyType int32  `json:"verify_type"`
}

type SetGroupHisMsgVisibleReq struct {
	GroupId            string `json:"group_id"`
	GroupHisMsgVisible int32  `json:"group_his_msg_visible"`
}

type GroupAdministratorsReq struct {
	GroupId  string   `json:"group_id"`
	AdminIds []string `json:"admin_ids"`
}

type GroupAdministratorsResp struct {
	GroupId string             `json:"group_id"`
	Items   []*GroupMemberInfo `json:"items"`
}

type SetGroupDisplayNameReq struct {
	GroupId        string `json:"group_id"`
	GrpDisplayName string `json:"grp_display_name"`
}

type QryGrpApplicationsResp struct {
	Items []*GrpApplicationItem `json:"items"`
}

type GrpApplicationItem struct {
	ApplicationId string   `json:"application_id"`
	GrpInfo       *GrpInfo `json:"grp_info"`
	ApplyType     int32    `json:"apply_type"`
	Sponsor       *UserObj `json:"sponsor"`
	Recipient     *UserObj `json:"recipient"`
	Inviter       *UserObj `json:"inviter"`
	Operator      *UserObj `json:"operator"`
	ApplyTime     int64    `json:"apply_time"`
	Status        int32    `json:"status"`
}

type GrpInfo struct {
	GroupId         string             `json:"group_id"`
	GroupName       string             `json:"group_name"`
	GroupPortrait   string             `json:"group_portrait"`
	Members         []*GroupMemberInfo `json:"members"`
	MemberCount     int32              `json:"member_count"`
	Owner           *GroupMemberInfo   `json:"owner"`
	MyRole          GrpMemberRole      `json:"my_role"`
	GroupManagement *GroupManagement   `json:"group_management"`
	GrpDisplayName  string             `json:"grp_display_name"`
	MemberOffset    string             `json:"member_offset"`
}

const (
	AttItemKey_HideGrpMsg      string = "hide_grp_msg"
	AttItemKey_GrpAnnouncement string = "grp_announcement"
	AttItemKey_GrpVerifyType   string = "grp_verify_type"
	AttItemKey_GrpDisplayName  string = "grp_display_name"

	AttItemKey_GrpEditMsgRight  string = "grp_edit_msg_right"
	AttItemKey_AddMemberRight   string = "grp_add_member_right"
	AttItemKey_MentionAllRight  string = "grp_mention_all_right"
	AttItemKey_TopMsgRight      string = "grp_top_msg_right"
	AttItemKey_SendMsgRight     string = "grp_send_msg_right"
	AttItemKey_SetMsgLifeRight  string = "grp_set_msg_life_right"
	AttItemKey_ApplyFriendRight string = "grp_apply_friend_right"
)
