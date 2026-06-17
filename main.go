package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/log"
	"github.com/juggleim/jugglechat-server/routers"
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
	routers.Route(httpServer, "jim")
	routers.LoadWebIM(httpServer)
	go httpServer.Run(fmt.Sprintf(":%d", configures.Config.Port))
	if configures.Config.CallbackPort > 0 {
		callbackServer := gin.Default()
		routers.CallbackRoute(callbackServer, "callback")
		go callbackServer.Run(fmt.Sprintf(":%d", configures.Config.CallbackPort))
	}

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
