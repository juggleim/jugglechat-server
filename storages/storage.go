package storages

import (
	"github.com/juggleim/jugglechat-server/storages/dbs"
	"github.com/juggleim/jugglechat-server/storages/models"
)

func NewUserStorage() models.IUserStorage {
	return &dbs.UserDao{}
}

func NewUserExtStorage() models.IUserExtStorage {
	return &dbs.UserExtDao{}
}

func NewFriendRelStorage() models.IFriendRelStorage {
	return &dbs.FriendRelDao{}
}

func NewFriendApplicationStorage() models.IFriendApplicationStorage {
	return &dbs.FriendApplicationDao{}
}

func NewGrpApplicationStorage() models.IGrpApplicationStorage {
	return &dbs.GrpApplicationDao{}
}

func NewQrCodeRecordStorage() models.IQrCodeRecordStorage {
	return &dbs.QrCodeRecordDao{}
}

func NewSmsRecordStorage() models.ISmsRecordStorage {
	return &dbs.SmsRecordDao{}
}

func NewTeleBotStorage() models.ITeleBotStorage {
	return &dbs.TeleBotDao{}
}

func NewGroupStorage() models.IGroupStorage {
	return &dbs.GroupDao{}
}

func NewGroupExtStorage() models.IGroupExtStorage {
	return &dbs.GroupExtDao{}
}

func NewGroupMemberStorage() models.IGroupMemberStorage {
	return &dbs.GroupMemberDao{}
}

func NewGroupAdminStorage() models.IGroupAdminStorage {
	return &dbs.GroupAdminDao{}
}

func NewFeedbackStorage() models.IFeedbackStorage {
	return &dbs.FeedbackDao{}
}

func NewConverConfStorage() models.IConverConfStorage {
	return &dbs.ConverConfDao{}
}

func NewApplicationStorage() models.IApplicationStorage {
	return &dbs.ApplicationDao{}
}

func NewBanUserStorage() models.IBanUserStorage {
	return &dbs.BanUserDao{}
}

func NewBlockUserStorage() models.IBlockUserStorage {
	return &dbs.BlockDao{}
}

func NewSensitiveWordStorage() models.ISensitiveWordStorage {
	return &dbs.SensitiveWordDao{}
}
