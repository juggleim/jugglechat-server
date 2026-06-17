package apis

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AssistantMsgCallback(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Printf("read assistant msg callback body error: %v\n", err)
		ctx.String(http.StatusOK, "ok")
		return
	}
	fmt.Printf("assistant msg callback body: %s\n", string(body))
	ctx.String(http.StatusOK, "ok")
}
