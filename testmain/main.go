package main

import (
	"fmt"

	"github.com/juggleim/commons/configures"
	"github.com/juggleim/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/log"
	"github.com/juggleim/jugglechat-server/storages/dbs"
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

	dao := dbs.UserDao{}
	users, err := dao.QryUsers("appkey", "ser", 0, 10, false)
	fmt.Println(err, users)
}
