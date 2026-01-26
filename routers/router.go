package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis"
)

func Route(eng *gin.Engine, prefix string) *gin.RouterGroup {
	eng.Use(corsHandler())
	group := eng.Group("/" + prefix)
	group.Use(apis.Validate)

	group.POST("/login", apis.Login)
	group.POST("/register", apis.Register)
	group.GET("/login/qrcode", apis.GenerateQrCode)
	group.POST("/login/qrcode/check", apis.CheckQrCode)
	group.POST("/login/qrcode/confirm", apis.ConfirmQrCode)
	group.POST("/sms/send", apis.SmsSend)
	group.POST("/sms_login", apis.SmsLogin)
	group.POST("/sms/login", apis.SmsLogin)
	group.POST("/email/send", apis.EmailSend)
	group.POST("/email/login", apis.EmailLogin)
	group.POST("/file_cred", apis.GetFileCred)
	group.POST("/translate", apis.Translate)
	group.GET("/syncconfs", apis.SyncConfs)

	group.POST("/users/update", apis.UpdateUser)
	group.POST("/users/updpass", apis.UpdatePass)
	group.POST("/users/updsettings", apis.UpdateUserSettings)
	group.POST("/users/bindemail/send", apis.BindEmailSendEmail)
	group.POST("/users/bindemail", apis.BindEmail)
	group.POST("/users/bindphone/send", apis.BindPhoneSendSms)
	group.POST("/users/bindphone", apis.BindPhone)
	group.POST("/users/onlinestatus", apis.QryUsersOnlineStauts)
	group.POST("/users/search", apis.SearchUsers)
	group.GET("/users/info", apis.QryUserInfo)
	group.GET("/users/qrcode", apis.QryUserQrCode)
	group.POST("/users/setaccount", apis.SetLoginAccount)
	group.POST("/users/blockusers/add", apis.BlockUsers)
	group.POST("/users/blockusers/del", apis.UnBlockUsers)
	group.GET("/users/blockusers/list", apis.QryBlockUsers)

	group.POST("/telegrambots/add", apis.TelegramBotAdd)
	group.POST("/telegrambots/del", apis.TelegramBotDel)
	group.POST("/telegrambots/batchdel", apis.TelegramBotBatchDel)
	group.GET("/telegrambots/list", apis.TelegramBotList)

	group.POST("/groups/add", apis.CreateGroup)
	group.POST("/groups/create", apis.CreateGroup)
	group.POST("/groups/update", apis.UpdateGroup)
	group.POST("/groups/dissolve", apis.DissolveGroup)
	group.POST("/groups/members/add", apis.AddGrpMembers)
	group.POST("/groups/apply", apis.GroupApply)
	group.POST("/groups/invite", apis.GroupInvite)
	group.POST("/groups/quit", apis.QuitGroup)
	group.POST("/groups/members/del", apis.DelGrpMembers)
	group.GET("/groups/members/list", apis.QryGrpMembers)
	group.POST("/groups/members/check", apis.CheckGroupMembers)
	group.POST("/groups/members/search", apis.SearchGroupMembers)
	group.GET("/groups/info", apis.QryGroupInfo)
	group.GET("/groups/qrcode", apis.QryGrpQrCode)
	group.POST("/groups/setgrpannouncement", apis.SetGrpAnnouncement)
	group.GET("/groups/getgrpannouncement", apis.GetGrpAnnouncement)
	group.POST("/groups/setdisplayname", apis.SetGrpDisplayName)
	//group manage
	group.POST("/groups/management/chgowner", apis.ChgGroupOwner)
	group.POST("/groups/management/administrators/add", apis.AddGrpAdministrator)
	group.POST("/groups/management/administrators/del", apis.DelGrpAdministrator)
	group.GET("/groups/management/administrators/list", apis.QryGrpAdministrators)
	group.POST("/groups/management/setmute", apis.SetGroupMute)
	group.POST("/groups/management/setgrpmembersmute", apis.SetGroupMembersMute)
	group.POST("/groups/management/setgrpverifytype", apis.SetGrpVerifyType)
	group.POST("/groups/management/sethismsgvisible", apis.SetGrpHisMsgVisible)
	group.POST("/groups/management/set", apis.SetGrpManagementConfs)
	group.GET("/groups/mygroups", apis.QryMyGroups)
	group.POST("/groups/mygroups/search", apis.SearchMyGroups)
	// grp application
	group.GET("/groups/myapplications", apis.QryMyGrpApplications)
	group.GET("/groups/mypendinginvitations", apis.QryMyPendingGrpInvitations)
	group.GET("/groups/grpinvitations", apis.QryGrpInvitations)
	group.GET("/groups/grppendingapplications", apis.QryGrpPendingApplications)
	group.GET("/groups/grpapplications", apis.QryGrpApplications)
	group.POST("/groups/grpapplications/confirm", apis.GroupComfirm)

	group.GET("/friends/list", apis.QryFriendsWithPage)
	group.POST("/friends/setdisplayname", apis.SetFriendDisplayName)
	group.POST("/friends/search", apis.SearchFriends)
	// group.POST("/friends/add", apis.AddFriend)
	group.POST("/friends/apply", apis.ApplyFriend)
	group.POST("/friends/confirm", apis.ConfirmFriend)
	group.POST("/friends/del", apis.DelFriend)
	group.GET("/friends/applications", apis.FriendApplications)
	group.GET("/friends/myapplications", apis.MyFriendApplications)
	group.GET("/friends/mypendingapplications", apis.MyPendingFriendApplications)

	//message operation
	group.POST("/messages/recall", apis.RecallMsg)
	group.POST("/messages/del", apis.DelMsgs)

	// conversation confs
	group.GET("/converconfs/get", apis.GetConverConfs)
	group.POST("/converconfs/set", apis.SetConverConfs)

	//dashboard applications
	group.GET("/applications/list", apis.QryApplications)

	//feedback
	group.POST("/feedbacks/add", apis.AddFeedback)
	return group
}

func corsHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Add("Access-Control-Allow-Headers", "*")
		context.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Writer.Header().Add("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}
