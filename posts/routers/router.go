package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server-post/apis"
)

func Route(group *gin.RouterGroup) {
	group.GET("/posts/list", apis.QryPosts)
	group.GET("/posts/info", apis.PostInfo)
	group.POST("/posts/add", apis.PostAdd)
	group.POST("/posts/update", apis.PostUpdate)
	group.POST("/posts/del", apis.PostDel)
	group.POST("/posts/reactions/add", apis.AddPostReaction)
	group.POST("/posts/reactions/del", apis.DelPostReaction)
	group.GET("/posts/reactions/list", apis.QryPostReactions)

	group.GET("/posts/comments/list", apis.QryPostComments)
	group.POST("/posts/comments/add", apis.PostCommentAdd)
	group.POST("/posts/comments/update", apis.PostCommentUpdate)
	group.POST("/posts/comments/del", apis.PostCommentDel)
}
