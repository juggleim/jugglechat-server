package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/services"
)

func SetPasswordBySMS(ctx *gin.Context) {
	var req struct {
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if e := ctx.ShouldBindJSON(&req); e != nil {
		responses.ErrorHttpResp(ctx, errs.IMErrorCode_APP_REQ_BODY_ILLEGAL)
		return
	}
	code := services.SetPasswordBySMS(ctxs.ToCtx(ctx), req.Code, req.NewPassword)
	if code != errs.IMErrorCode_SUCCESS {
		responses.ErrorHttpResp(ctx, code)
		return
	}
	responses.SuccessHttpResp(ctx, nil)
}
