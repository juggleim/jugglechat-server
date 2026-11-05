package apis

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/services/deviceauth"
)

func TestDeviceHandlersRequireVerifiedSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, handler := range []gin.HandlerFunc{ListDevices, DeleteDevice} {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		handler(ctx)
		var body struct {
			Code int `json:"code"`
		}
		if e := json.Unmarshal(w.Body.Bytes(), &body); e != nil || body.Code != 17605 {
			t.Fatalf("%s %v", w.Body.String(), e)
		}
	}
}
func TestLoginParameterValidationBeforeIssuance(t *testing.T) {
	old := configures.Config.DeviceAuthRequiredAt
	defer func() { configures.Config.DeviceAuthRequiredAt = old }()
	configures.Config.DeviceAuthRequiredAt = map[string]int64{"app": 1}
	for _, body := range []string{`{}`, `{"device_id":"d"}`, `{"device_id":null,"platform":"ios"}`, `{"device_id":"d","platform":"x"}`} {
		var p deviceauth.DeviceParams
		if e := json.Unmarshal([]byte(body), &p); e != nil {
			t.Fatal(e)
		}
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Set(string(ctxs.CtxKey_AppKey), "app")
		if _, e := issueLogin(ctx, "u", p); e == nil {
			t.Fatal("invalid params reached token issuance", body)
		}
	}
}
func TestDeleteDeviceRejectsMalformedJSON(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set("device_session", deviceauth.Claims{AppKey: "a", UserID: "u", DeviceID: "current", SessionID: "s", Kind: "user_session"})
	ctx.Request = httptest.NewRequest("POST", "/jim/devices/del", strings.NewReader(`{"device_id":2}`))
	DeleteDevice(ctx)
	if ctx.Writer.Status() != 200 {
		t.Fatal(ctx.Writer.Status())
	}
}
