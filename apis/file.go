package apis

import (
	"appserver/services"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetFileCred(ctx *gin.Context) {
	cred := services.GetFileCred(context.Background())
	ctx.JSON(http.StatusOK, services.SuccessResp(cred))
}
