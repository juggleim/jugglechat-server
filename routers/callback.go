package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis"
	"github.com/juggleim/jugglechat-server/services"
)

func CallbackRoute(eng *gin.Engine, prefix string) *gin.RouterGroup {
	services.CallbackUrlPrefix = prefix
	group := eng.Group("/" + prefix)

	group.POST("/assistants/msgcallback", apis.AssistantMsgCallback)

	return group
}
