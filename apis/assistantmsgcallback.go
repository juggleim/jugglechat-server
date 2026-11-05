package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
)

func AssistantMsgCallback(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		fmt.Printf("read assistant msg callback body error: %v\n", err)
		responses.SuccessHttpResp(ctx, nil)
		return
	}
	var req models.CustomChatMsgReq
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("unmarshal assistant msg callback body error: %v\n", err)
		responses.SuccessHttpResp(ctx, nil)
		return
	}
	token := services.GenerateToken(req.AppKey, req.Sender)
	if configures.Config.AssistantServiceCredentials {
		info, ok := appinfos.GetAppInfo(req.AppKey)
		if !ok {
			responses.SuccessHttpResp(ctx, nil)
			return
		}
		var err error
		token, err = deviceauth.SignService(req.AppKey, req.Sender, []byte(info.AppSecureKey), time.Now().UnixMilli())
		if err != nil {
			responses.SuccessHttpResp(ctx, nil)
			return
		}
	}
	req.Token = &token

	if configures.Config.AssistantAgentUrl != "" {
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		_, code, err := tools.HttpDo(http.MethodPost, configures.Config.AssistantAgentUrl, headers, tools.ToJson(req))
		if err != nil {
			fmt.Printf("post assistant msg callback to agent error: %v\n", err)
		} else {
			fmt.Printf("post assistant msg callback to agent response code: %d\n", code)
		}
	} else {
		fmt.Printf("assistant agent url is empty, skip post assistant msg callback\n")
	}
	responses.SuccessHttpResp(ctx, nil)
}
