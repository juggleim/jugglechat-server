package models

type FriendRel struct {
	ID          int64
	AppKey      string
	UserId      string
	FriendId    string
	OrderTag    string
	DisplayName string
}

type IFriendRelStorage interface {
	Upsert(item FriendRel) error
	BatchUpsert(items []FriendRel) error
	QueryFriendRels(appkey, userId string, startId, limit int64) ([]*FriendRel, error)
	QueryFriendRelsWithPage(appkey, userId string, orderTag string, page, size int64) ([]*User, error)
	SearchFriendsByName(appkey, userId string, nickname string, startId, limit int64) ([]*User, error)
	BatchDelete(appkey, userId string, friendIds []string) error
	QueryFriendRelsByFriendIds(appkey, userId string, friendIds []string) ([]*FriendRel, error)
	UpdateOrderTag(appkey, userId, friendId string, orderTag string) error
	UpdateDisplayName(appkey, userId, friendId string, displayName string) error
}

type FriendApplicationStatus int

var (
	FriendApplicationStatus_Apply   FriendApplicationStatus = 0
	FriendApplicationStatus_Agree   FriendApplicationStatus = 1
	FriendApplicationStatus_Decline FriendApplicationStatus = 2
	FriendApplicationStatus_Expired FriendApplicationStatus = 3
)

type FriendApplication struct {
	ID          int64
	RecipientId string
	SponsorId   string
	ApplyTime   int64
	Status      FriendApplicationStatus
	AppKey      string
}

type IFriendApplicationStorage interface {
	Upsert(item FriendApplication) error
	QueryPendingApplications(appkey, recipientId string, startTime, count int64, isPositive bool) ([]*FriendApplication, error)
	QueryMyApplications(appkey, sponsorId string, startTime, count int64, isPositive bool) ([]*FriendApplication, error)
	QueryApplications(appkey, userId string, startTime, count int64, isPositive bool) ([]*FriendApplication, error)
	UpdateStatus(appkey, sponsorId, recipientId string, status FriendApplicationStatus) error
}
