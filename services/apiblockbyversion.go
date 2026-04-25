package services

import (
	"context"
	"strconv"
	"strings"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
)

func CheckApiBlockByVersion(ctx context.Context) bool {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	if appkey != "" {
		appinfo, exist := appinfos.GetAppInfo(appkey)
		if exist && appinfo != nil {
			if exist, obj := appinfo.GetExt("app_block_version"); exist && obj != nil {
				blockVersion := obj.(string)
				blockVersion = strings.TrimSpace(blockVersion)
				version := strings.TrimSpace(ctxs.GetVersionFromCtx(ctx))
				if version == "" {
					version = "0"
				}
				if blockVersion == "" {
					return false
				}
				return compareVersion(version, blockVersion) < 0
			}
		}
	}
	return false
}

func compareVersion(version, blockVersion string) int {
	left := strings.Split(version, ".")
	right := strings.Split(blockVersion, ".")
	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}
	for i := 0; i < maxLen; i++ {
		lv := 0
		rv := 0
		if i < len(left) {
			lv, _ = strconv.Atoi(left[i])
		}
		if i < len(right) {
			rv, _ = strconv.Atoi(right[i])
		}
		if lv < rv {
			return -1
		}
		if lv > rv {
			return 1
		}
	}
	return 0
}
