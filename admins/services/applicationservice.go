package services

import (
	"context"
	"fmt"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/admins/apis/models"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"
)

func AddApplication(ctx context.Context, application *apimodels.Application) (errs.AdminErrorCode, *apimodels.Application) {
	if application.AppKey == "" {
		return errs.AdminErrorCode_AppNotExist, nil
	}
	storage := storages.NewApplicationStorage()
	appId := tools.GenerateUUIDShort11()
	err := storage.Create(models.Application{
		AppId:    appId,
		AppName:  application.AppName,
		AppIcon:  application.AppIcon,
		AppDesc:  application.AppDesc,
		AppUrl:   application.AppUrl,
		AppOrder: application.AppOrder,
		AppKey:   application.AppKey,
	})
	if err != nil {
		return errs.AdminErrorCode_ServerErr, nil
	}
	return errs.AdminErrorCode_Success, &apimodels.Application{
		AppId:       appId,
		AppName:     application.AppName,
		AppIcon:     application.AppIcon,
		AppDesc:     application.AppDesc,
		AppUrl:      application.AppUrl,
		AppOrder:    application.AppOrder,
		CreatedTime: time.Now().UnixMilli(),
		UpdatedTime: time.Now().UnixMilli(),
		AppKey:      application.AppKey,
	}
}

func UpdApplication(ctx context.Context, application *apimodels.Application) errs.AdminErrorCode {
	storage := storages.NewApplicationStorage()
	err := storage.Update(models.Application{
		AppId:    application.AppId,
		AppName:  application.AppName,
		AppIcon:  application.AppIcon,
		AppDesc:  application.AppDesc,
		AppUrl:   application.AppUrl,
		AppOrder: application.AppOrder,
		AppKey:   application.AppKey,
	})
	if err != nil {
		fmt.Println("err:", err)
		return errs.AdminErrorCode_ServerErr
	}
	return errs.AdminErrorCode_Success
}

func DelApplications(ctx context.Context, appIds *apimodels.ApplicationIds) errs.AdminErrorCode {
	storage := storages.NewApplicationStorage()
	err := storage.BatchDelete(appIds.AppKey, appIds.AppIds)
	if err != nil {
		return errs.AdminErrorCode_ServerErr
	}
	return errs.AdminErrorCode_Success
}

func QryApplications(ctx context.Context, appkey string, page, size int64, isPositive bool) (errs.AdminErrorCode, *apimodels.Applications) {
	ret := &apimodels.Applications{
		Items: []*apimodels.Application{},
		Page:  int(page),
		Size:  int(size),
	}
	storage := storages.NewApplicationStorage()
	items, err := storage.QryApplicationsByPage(appkey, page, size)
	if err != nil {
		return errs.AdminErrorCode_ServerErr, ret
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
	return errs.AdminErrorCode_Success, ret
}
