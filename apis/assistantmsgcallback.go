package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"
)

func AssistantMsgCallback(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Printf("read assistant msg callback body error: %v\n", err)
		responses.SuccessHttpResp(ctx, nil)
		return
	}
	fmt.Printf("assistant msg callback body: %s\n", string(body))
	var req models.CustomChatMsgReq
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("unmarshal assistant msg callback body error: %v\n", err)
		responses.SuccessHttpResp(ctx, nil)
		return
	}
	token := services.GenerateToken(req.AppKey, req.Sender)
	req.Token = &token

	if configures.Config.AssistantAgentUrl != "" {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		resp, code, err := tools.HttpDo(http.MethodPost, configures.Config.AssistantAgentUrl, headers, tools.ToJson(req))
		if err != nil {
			fmt.Printf("post assistant msg callback to agent error: %v\n", err)
		} else {
			fmt.Printf("post assistant msg callback to agent response code: %d, body: %s\n", code, resp)
		}
	} else {
		fmt.Printf("assistant agent url is empty, skip post assistant msg callback\n")
	}
	responses.SuccessHttpResp(ctx, nil)
}
