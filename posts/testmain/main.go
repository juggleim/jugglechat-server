package main

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/util/log"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/posts/storages/dbs"
)

func main() {
	//init configure
	if err := configures.InitConfigures(); err != nil {
		fmt.Println("Init Configures failed", err)
		return
	}
	//init mysql
	if err := dbcommons.InitMysql(); err != nil {
		log.Error("Init Mysql failed.", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxs.CtxKey_AppKey, "appkey")
	ctx = context.WithValue(ctx, ctxs.CtxKey_RequesterId, "userid2")

	// code := services.AddPostReaction(ctx, &models.PostReaction{
	// 	PostId: "post1",
	// 	Key:    "k3",
	// 	Value:  "v3",
	// })

	// code := services.DelPostReaction(ctx, &models.PostReaction{
	// 	PostId: "post1",
	// 	Key:    "k1",
	// })
	// val := services.TopReactions(ctx, "post1", 10)

	dao := dbs.PostCommentFeedDao{}
	items, err := dao.QryPostComments("appkey", "userid2", "post1", 0, 10, false)
	if err == nil {
		fmt.Println(tools.ToJson(items))
	} else {
		fmt.Println("xxx:", err)
	}

	// fmt.Println(code)
}
