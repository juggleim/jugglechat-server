package services

import (
	"context"

	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/sensitive"
)

func CheckSensitiveText(ctx context.Context, text string) (bool, string) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	sensitiveService := sensitive.GetAppSensitiveFilter(appkey)
	if sensitiveService != nil {
		isDeny, newText := sensitiveService.ReplaceSensitiveWords(text)
		if isDeny {
			return false, ""
		} else if text != newText {
			return true, newText
		}
	}
	return true, text
}
