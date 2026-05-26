package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	adminRouters "github.com/juggleim/jugglechat-server/admins/routers"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/log"
	"github.com/juggleim/jugglechat-server/routers"

	aiRouters "github.com/juggleim/jugglechat-server-ai/routers"
)

func main() {
	//init configure
	if err := configures.InitConfigures(); err != nil {
		fmt.Println("Init Configures failed", err)
		return
	}
	//init log
	log.InitLogs()
	//init mysql
	if err := dbcommons.InitMysql(); err != nil {
		log.Error("Init Mysql failed.", err)
		return
	}
	//upgrade db
	dbcommons.Upgrade()

	httpServer := gin.Default()
	grp := routers.Route(httpServer, "jim")
	aiRouters.Route(grp)
	go httpServer.Run(fmt.Sprintf(":%d", configures.Config.Port))

	//start admin
	adminServer := gin.Default()
	group := adminRouters.RouteLogin(adminServer, "admingateway")
	adminRouters.RouteProxy(group)
	adminRouters.Route(group)
	adminRouters.LoadJuggleChatAdminWeb(adminServer)
	go adminServer.Run(fmt.Sprintf(":%d", configures.Config.AdminPort))

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigChan
		signal.Stop(sigChan)
		close(closeChan)
	}()

	<-closeChan
}
