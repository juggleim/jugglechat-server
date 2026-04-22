package main

import (
	"fmt"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/log"
	"github.com/juggleim/jugglechat-server/storages"
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

	storage := storages.NewGroupMemberStorage()

	members, err := storage.QueryMembers("appkey", "userid1", "groupid1", 0, 2)
	fmt.Println(err)
	fmt.Println(len(members))
}
