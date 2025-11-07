package services

import (
	"context"

	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/commons/errs"
)

func SetEmailConf(ctx context.Context, req *apimodels.EmailConf) errs.AdminErrorCode {
	return errs.AdminErrorCode_Success
}

func GetEmailConf(ctx context.Context, appkey string) (errs.AdminErrorCode, *apimodels.EmailConf) {
	return errs.AdminErrorCode_Success, nil
}
