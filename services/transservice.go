package services

import (
	"context"
	"sync"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/appinfos"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/commons/transengines"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

func Translate(ctx context.Context, req *apimodels.TransReq) (errs.IMErrorCode, *apimodels.TransReq) {
	if req.TargetLang == "" || len(req.Items) <= 0 {
		return errs.IMErrorCode_APP_REQ_BODY_ILLEGAL, nil
	}
	resp := &apimodels.TransReq{
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
		Items:      []*apimodels.TransItem{},
	}
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	transEngine := GetTransEngine(appkey)
	if transEngine != nil && transEngine != transengines.DefaultTransEngine {
		wg := &sync.WaitGroup{}
		for _, item := range req.Items {
			afterItem := &apimodels.TransItem{
				Key:     item.Key,
				Content: item.Content,
			}
			resp.Items = append(resp.Items, afterItem)
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := transEngine.Translate(afterItem.Content, []string{req.TargetLang})
				if len(result) > 0 {
					if afterTranslated, exist := result[req.TargetLang]; exist {
						afterItem.Content = afterTranslated
					}
				}
			}()
		}
		wg.Wait()
	} else {
		return errs.IMErrorCode_APP_TRANS_NOTRANSENGINE, nil
	}
	return errs.IMErrorCode_SUCCESS, resp
}

func GetTransEngine(appkey string) transengines.ITransEngine {
	appInfo, exist := appinfos.GetAppInfo(appkey)
	if exist && appInfo != nil {
		if appInfo.TransEngine == nil {
			appinfos.GetAppLock().Lock()
			defer appinfos.GetAppLock().Unlock()
			loadTransEngine(appInfo)
		}
		if appInfo.TransEngine != nil {
			return appInfo.TransEngine
		}
	}
	return transengines.DefaultTransEngine
}

func loadTransEngine(appInfo *appinfos.AppInfo) {
	extDao := dbs.AppExtDao{}
	ext, err := extDao.Find(appInfo.AppKey, "trans_engine_conf")
	if err == nil && ext.AppItemValue != "" {
		transConf := &TransEngineConf{}
		err = utils.JsonUnMarshal([]byte(ext.AppItemValue), transConf)
		if err == nil {
			if transConf.Channel == "baidu" && transConf.BdTransEngine != nil && transConf.BdTransEngine.ApiKey != "" && transConf.BdTransEngine.SecretKey != "" {
				appInfo.TransEngine = &transengines.BdTransEngine{
					AppKey:    appInfo.AppKey,
					ApiKey:    transConf.BdTransEngine.ApiKey,
					SecretKey: transConf.BdTransEngine.SecretKey,
				}
			} else if transConf.Channel == "deepl" && transConf.DeeplTransEngine != nil && transConf.DeeplTransEngine.AuthKey != "" {
				appInfo.TransEngine = &transengines.DeeplTransEngine{
					AppKey:  appInfo.AppKey,
					AuthKey: transConf.DeeplTransEngine.AuthKey,
				}
			} else {
				if transConf.BdTransEngine != nil && transConf.BdTransEngine.ApiKey != "" && transConf.BdTransEngine.SecretKey != "" {
					appInfo.TransEngine = &transengines.BdTransEngine{
						AppKey:    appInfo.AppKey,
						ApiKey:    transConf.BdTransEngine.ApiKey,
						SecretKey: transConf.BdTransEngine.SecretKey,
					}
				} else if transConf.DeeplTransEngine != nil && transConf.DeeplTransEngine.AuthKey != "" {
					appInfo.TransEngine = &transengines.DeeplTransEngine{
						AppKey:  appInfo.AppKey,
						AuthKey: transConf.DeeplTransEngine.AuthKey,
					}
				} else {
					appInfo.TransEngine = transengines.DefaultTransEngine
				}
			}
		}
	}
}

type TransEngineConf struct {
	Channel          string                         `json:"channel,omitempty"`
	BdTransEngine    *transengines.BdTransEngine    `json:"baidu,omitempty"`
	DeeplTransEngine *transengines.DeeplTransEngine `json:"deepl,omitempty"`
}
