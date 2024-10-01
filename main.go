package main

import (
	"appserver/apis"
	"appserver/configures"
	"appserver/dbs"
	"appserver/log"
	"appserver/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	//init configure
	if err := configures.InitConfigures(); err != nil {
		fmt.Println("Init Configures failed", err)
		return
	}
	//init log
	log.InitLogs()
	//init sms
	if err := services.InitSms(); err != nil {
		log.Error("Init Sms failed", err)
		return
	}
	//init im sdk
	if err := services.InitImSdk(); err != nil {
		log.Error("Init Im Sdk failed", err)
		return
	}
	//init mysql
	if err := dbs.InitMysql(); err != nil {
		log.Error("Init Mysql failed.", err)
		return
	}

	server := gin.Default()
	server.Use(CorsHandler())
	group := server.Group("/")
	group.Use(apis.Validate)
	group.POST("sms/send", apis.SmsSend)
	group.POST("/sms_login", apis.SmsLogin)
	group.GET("/file_cred", apis.GetFileCred)

	//user
	group.POST("/users/update", apis.UpdateUser)
	group.POST("/users/search", apis.SearchByPhone)
	group.GET("/users/info", apis.QryUserInfo)

	//group
	group.POST("/groups/add", apis.CreateGroup)
	group.POST("/groups/update", apis.UpdateGroup)
	group.POST("/groups/members/add", apis.AddGrpMembers)
	group.POST("/groups/members/del", apis.DelGrpMembers)
	group.GET("/groups/info", apis.QryGroup)
	group.GET("/groups/mygroups", apis.QryMyGroups)

	//friend
	group.GET("/friends/list", apis.QryFriends)
	group.POST("/friends/add", apis.AddFriend)

	fmt.Println("Start Server with port:", configures.Config.Port)
	server.Run(fmt.Sprintf(":%d", configures.Config.Port))

}
func CorsHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Add("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization")
		context.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Writer.Header().Add("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}
