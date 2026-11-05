package apis

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/commons/responses"
	"github.com/juggleim/jugglechat-server/services"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
)

func deviceStore() deviceauth.Store {
	return deviceauth.Store{DB: dbcommons.GetDb(), IM: deviceauth.ConfiguredIM()}
}
func deviceError(ctx *gin.Context, err error, data interface{}) {
	code := 17600
	msg := "device operation failed; retry later"
	switch {
	case errors.Is(err, deviceauth.ErrQRExpired):
		code = 17007
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrParams):
		code = 17601
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrNotFound):
		code = 17602
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrCurrent):
		code = 17603
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrAuth):
		code = 17604
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrUpgrade):
		code = 17605
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrIM):
		code = 17606
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrQR):
		code = 17608
		msg = err.Error()
	case errors.Is(err, deviceauth.ErrPending):
		code = 17006
		msg = err.Error()
	}
	ctx.JSON(200, gin.H{"code": code, "msg": msg, "data": data})
}
func issueLogin(ctx *gin.Context, user string, p deviceauth.DeviceParams) (string, error) {
	app := ctx.GetString(string(ctxs.CtxKey_AppKey))
	now := time.Now().UnixMilli()
	legacy, e := p.Validate(deviceauth.Required(app, now))
	if e != nil {
		return "", e
	}
	info, ok := appinfos.GetAppInfo(app)
	if !ok {
		return "", deviceauth.ErrAuth
	}
	if legacy {
		deviceauth.LegacyLogins.Add(1)
		return legacyLoginToken(app, user, []byte(info.AppSecureKey), now)
	}
	return deviceStore().Issue(ctx.Request.Context(), app, user, p, []byte(info.AppSecureKey), now)
}
func legacyLoginToken(app, user string, key []byte, now int64) (string, error) {
	return (services.ImToken{AppKey: app, UserId: user, TokenTime: now}).ToTokenString(key)
}
func currentClaims(ctx *gin.Context) (deviceauth.Claims, bool) {
	v, ok := ctx.Get("device_session")
	if !ok {
		return deviceauth.Claims{}, false
	}
	c, ok := v.(deviceauth.Claims)
	return c, ok
}
func ListDevices(ctx *gin.Context) {
	c, ok := currentClaims(ctx)
	if !ok {
		deviceError(ctx, deviceauth.ErrUpgrade, nil)
		return
	}
	g, e := deviceStore().List(ctx.Request.Context(), c)
	if e != nil {
		deviceError(ctx, e, nil)
		return
	}
	responses.SuccessHttpResp(ctx, g)
}
func DeleteDevice(ctx *gin.Context) {
	c, ok := currentClaims(ctx)
	if !ok {
		deviceError(ctx, deviceauth.ErrUpgrade, nil)
		return
	}
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if ctx.ShouldBindJSON(&req) != nil {
		deviceError(ctx, deviceauth.ErrParams, nil)
		return
	}
	j, e := deviceStore().Delete(ctx.Request.Context(), c, req.DeviceID, time.Now().UnixMilli())
	status := "pending"
	if j.Done {
		status = "completed"
	}
	if j.Done && j.LastError == "superseded" {
		status = "superseded"
	}
	data := gin.H{"operation_id": j.OperationID, "status": status}
	if e != nil {
		deviceError(ctx, e, data)
		return
	}
	responses.SuccessHttpResp(ctx, data)
}
