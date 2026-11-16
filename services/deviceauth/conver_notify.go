package deviceauth

import (
	"context"
	"net/http"

	juggleimsdk "github.com/juggleim/imserver-sdk-go"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

// NotifyConverConfig uses the verified SDK message schema, but not its
// context-free HTTP transport. MsgId is deliberately unset: no idempotency
// contract has been verified. Subchannel routing is not supported by this SDK.
func (a *HTTPIM) NotifyConverConfig(ctx context.Context, j dbs.ConverOutbox) error {
	ctx, cancel := a.bounded(ctx)
	defer cancel()
	c, e := a.config(ctx, j.AppKey)
	if e != nil {
		return e
	}
	if j.ActorID == "" || j.TargetID == "" || (j.ConverType != 1 && j.ConverType != 2) {
		return imError("invalid notification scope")
	}
	yes, no := true, false
	msg := juggleimsdk.Message{SenderId: j.ActorID, TargetId: j.TargetID, MsgType: "jgd:lifetime_msg", MsgContent: j.Payload, IsStorage: &yes, IsCount: &no, IsNotifySender: &yes, IsCmd: &yes}
	path := "/apigateway/messages/private/send"
	if j.ConverType == 2 {
		path = "/apigateway/messages/group/send"
	}
	return a.call(ctx, c, http.MethodPost, path, msg, nil)
}
