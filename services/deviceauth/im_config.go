package deviceauth

import (
	"context"

	"github.com/juggleim/jugglechat-server/commons/configures"
	"github.com/juggleim/jugglechat-server/commons/dbcommons"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

// ConfiguredIM uses the existing IM origin and app credentials. The context-aware
// lookup avoids the context-free app cache miss path for HTTP deadlines.
func ConfiguredIM() IMAdapter {
	return &HTTPIM{Resolve: func(ctx context.Context, app string) (IMConfig, error) {
		var info dbs.AppInfoDao
		db := dbcommons.GetDb()
		if db == nil {
			return IMConfig{}, imError("app configuration unavailable")
		}
		if e := db.WithContext(ctx).Select("app_key", "app_secret", "app_status").Where("app_key=?", app).Take(&info).Error; e != nil || info.AppKey != app || info.AppStatus != 0 {
			return IMConfig{}, imError("app configuration unavailable")
		}
		return IMConfig{BaseURL: configures.Config.ImApiDomain, AppKey: info.AppKey, Secret: info.AppSecret, DeviceKick: configures.Config.DeviceKickEnabled()}, nil
	}}
}
