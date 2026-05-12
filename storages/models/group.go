package models

import "time"

type GrpApplicationStatus int

var (
	GrpApplicationStatus_Apply        GrpApplicationStatus = 0
	GrpApplicationStatus_AgreeApply   GrpApplicationStatus = 1
	GrpApplicationStatus_DeclineApply GrpApplicationStatus = 2
	GrpApplicationStatus_ExpiredApply GrpApplicationStatus = 3

	GrpApplicationStatus_Invite        GrpApplicationStatus = 10
	GrpApplicationStatus_AgreeInvite   GrpApplicationStatus = 11
	GrpApplicationStatus_DeclineInvite GrpApplicationStatus = 12
	GrpApplicationStatus_ExpiredInvite GrpApplicationStatus = 13
)

type GrpApplicationType int

var (
	GrpApplicationType_Invite GrpApplicationType = 0
	GrpApplicationType_Apply  GrpApplicationType = 1
)

type GrpApplication struct {
	ID          int64
	GroupId     string
	ApplyType   GrpApplicationType
	SponsorId   string
	RecipientId string
	InviterId   string
	OperatorId  string
	ApplyTime   int64
	Status      GrpApplicationStatus
	AppKey      string
}

type IGrpApplicationStorage interface {
	FindById(appkey string, id int64) (*GrpApplication, error)
	InviteUpsert(item GrpApplication) error
	ApplyUpsert(item GrpApplication) error
	UpdateStatus(id int64, status GrpApplicationStatus) error
	QueryMyGrpApplications(appkey, sponsorId string, startTime, count int64, isPositive bool) ([]*GrpApplication, error)
	QueryMyPendingGrpInvitations(appkey, recipientId string, startTime, count int64, isPositive bool) ([]*GrpApplication, error)
	QueryGrpInvitations(appkey, groupId string, startTime, count int64, isPositive bool) ([]*GrpApplication, error)
	QueryGrpPendingApplications(appkey, groupId string, startTime, count int64, isPositive bool) ([]*GrpApplication, error)
	QueryGrpApplications(appkey, groupId string, startTime, count int64, isPositive bool) ([]*GrpApplication, error)
}

type Group struct {
	ID            int64
	GroupId       string
	GroupName     string
	GroupPortrait string
	CreatorId     string
	CreatedTime   time.Time
	UpdatedTime   time.Time
	AppKey        string
	IsMute        int
}

type IGroupStorage interface {
	Create(item Group) error
	IsExist(appkey, groupId string) (bool, error)
	FindById(appkey, groupId string) (*Group, error)
	Delete(appkey, groupId string) error
	UpdateGroupMuteStatus(appkey, groupId string, isMute int32) error
	UpdateGrpName(appkey, groupId, groupName, groupPortrait string) error
	UpdateCreatorId(appkey, groupId, creatorId string) error
	QryGroups(appkey, name string, startId, limit int64, isPositive bool) ([]*Group, error)
}

type GroupExt struct {
	ID          int64
	GroupId     string
	ItemKey     string
	ItemValue   string
	ItemType    int
	UpdatedTime time.Time
	AppKey      string
}

type IGroupExtStorage interface {
	Upsert(item GroupExt) error
	BatchUpsert(items []GroupExt) error
	Find(appkey, groupId string, itemKey string) (*GroupExt, error)
	QryExtFields(appkey, groupId string) ([]*GroupExt, error)
}

type GroupMember struct {
	ID             int64
	GroupId        string
	GroupName      string
	GroupPortrait  string
	MemberId       string
	Nickname       string
	UserPortrait   string
	MemberType     int
	CreatedTime    time.Time
	AppKey         string
	IsMute         int
	IsAllow        int
	MuteEndAt      int64
	GrpDisplayName string

	MemberFriendInfo *FriendInfo
}

type IGroupMemberStorage interface {
	Create(item GroupMember) error
	Find(appkey, groupId, memberId string) (*GroupMember, error)
	FindByMemberIds(appkey, groupId string, memberIds []string) ([]*GroupMember, error)
	BatchCreate(items []GroupMember) error
	QueryMembers(appkey, userId, groupId string, startId, limit int64) ([]*GroupMember, error)
	SearchMembersByName(appkey, userId, groupId, nickname string, startId, limit int64) ([]*GroupMember, error)
	QueryGroupsByMemberId(appkey, memberId string, startId, limit int64) ([]*GroupMember, error)
	BatchDelete(appkey, groupId string, memberIds []string) error
	DeleteByGroupId(appkey, groupId string) error
	UpdateMute(appkey, groupId string, isMute int, memberIds []string, muteEndAt int64) error
	UpdateAllow(appkey, groupId string, isAllow int, memberIds []string) error
	CountByGroup(appkey, groupId string) int
	UpdateGrpDisplayName(appkey, groupId, memberId string, displayName string) error
}

type GroupAdmin struct {
	ID             int64
	GroupId        string
	AdminId        string
	Nickname       string
	UserPortrait   string
	UserType       int
	GrpDisplayName string
	CreatedTime    time.Time
	AppKey         string
}

type IGroupAdminStorage interface {
	Upsert(item GroupAdmin) error
	QryAdmins(appkey, groupId string) ([]*GroupAdmin, error)
	BatchDel(appkey, groupId string, adminIds []string) error
	CheckAdmin(appkey, groupId, userId string) bool
}
