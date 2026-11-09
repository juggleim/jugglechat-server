package services

import (
	"context"
	"regexp"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/storages"
)

// Format eligibility only; this does not verify reachability or SMS capability.
var passwordStatusPhone = regexp.MustCompile(`^(1[3-9][0-9]{9}|\+[1-9][0-9]{6,14})$`)

func QryPasswordStatus(ctx context.Context) (errs.IMErrorCode, *apimodels.PasswordStatus) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	requester := ctxs.GetRequesterIdFromCtx(ctx)
	if appkey == "" {
		return errs.IMErrorCode_APP_NOT_EXISTED, nil
	}
	kind, _ := ctx.Value(ctxs.CtxKey_AuthKind).(string)
	if requester == "" || (kind != "legacy" && kind != "user_session") {
		return errs.IMErrorCode_APP_NOT_LOGIN, nil
	}
	state, err := storages.NewUserStorage().FindPasswordState(appkey, requester)
	if err != nil {
		return errs.IMErrorCode_APP_DEFAULT, nil
	}
	if state == nil {
		return errs.IMErrorCode_APP_USER_NOT_EXIST, nil
	}
	hasPassword := state.LoginPass != ""
	return errs.IMErrorCode_SUCCESS, &apimodels.PasswordStatus{
		HasPassword:         hasPassword,
		CanSetPasswordBySMS: !hasPassword && passwordStatusPhone.MatchString(state.Phone),
	}
}
