package services

import (
	"context"
	"testing"

	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
)

// Rejected identities must stop before any storage access (DB is uninitialized).
func TestPasswordStatusRequiresUserIdentity(t *testing.T) {
	for _, tc := range []struct {
		app, requester, kind string
		want                 errs.IMErrorCode
	}{
		{"", "u1", "legacy", 17002},
		{"app", "", "legacy", 17005},
		{"app", "", "user_session", 17005},
		{"app", "u1", "", 17005},
		{"app", "u1", "api_key", 17005},
		{"app", "u1", "service", 17005},
	} {
		ctx := context.WithValue(context.Background(), ctxs.CtxKey_AppKey, tc.app)
		ctx = context.WithValue(ctx, ctxs.CtxKey_RequesterId, tc.requester)
		ctx = context.WithValue(ctx, ctxs.CtxKey_AuthKind, tc.kind)
		code, data := QryPasswordStatus(ctx)
		if code != tc.want || data != nil {
			t.Errorf("%+v: code=%d data=%+v", tc, code, data)
		}
	}
}
