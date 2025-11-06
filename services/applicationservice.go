package services

import (
	"context"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/storages"
)

func QryApplications(ctx context.Context, page, size int64, isPositiveOrder bool) (errs.IMErrorCode, *apimodels.Applications) {
	ret := &apimodels.Applications{
		Items: []*apimodels.Application{},
		Page:  int(page),
		Size:  int(size),
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	storage := storages.NewApplicationStorage()
	items, err := storage.QryApplicationsByPage(appkey, page, size)
	if err != nil {
		return errs.IMErrorCode(errs.AdminErrorCode_ServerErr), ret
	}
	for _, item := range items {
		ret.Items = append(ret.Items, &apimodels.Application{
			AppId:       item.AppId,
			AppName:     item.AppName,
			AppIcon:     item.AppIcon,
			AppDesc:     item.AppDesc,
			AppUrl:      item.AppUrl,
			AppOrder:    item.AppOrder,
			CreatedTime: item.CreatedTime,
			UpdatedTime: item.UpdatedTime,
		})
	}
	return errs.IMErrorCode_SUCCESS, ret
}
