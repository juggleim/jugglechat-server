package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	apimodels "github.com/juggleim/jugglechat-server/apis/models"
	"github.com/juggleim/jugglechat-server/commons/caches"
	"github.com/juggleim/jugglechat-server/commons/ctxs"
	"github.com/juggleim/jugglechat-server/commons/errs"
	"github.com/juggleim/jugglechat-server/commons/fileengines"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/storages/dbs"
)

func GetFileCred(ctx context.Context, req *apimodels.QryFileCredReq) (errs.IMErrorCode, *apimodels.QryFileCredResp) {
	appkey := ctxs.GetAppKeyFromCtx(ctx)
	fileConf := GetFileConf(ctx, appkey)
	if fileConf == nil || fileConf == notExistFileConf {
		return errs.IMErrorCode_APP_FILE_NOOSS, nil
	}
	dir := fileTypeToDir(req.FileType)
	switch fileConf.FileEngine {
	case fileengines.ChannelQiNiu:
		if fileConf.QiNiu == nil {
			return errs.IMErrorCode_APP_FILE_NOOSS, nil
		}
		uploadToken, domain := fileConf.QiNiu.UploadToken(req.Ext)
		return errs.IMErrorCode_SUCCESS, &apimodels.QryFileCredResp{
			OssType: apimodels.OssType_QiNiu,
			QiNiuCredResp: &apimodels.QiNiuCredResp{
				Domain: domain,
				Token:  uploadToken,
			},
		}
	case fileengines.ChannelMinio:
		if fileConf.Minio == nil {
			return errs.IMErrorCode_APP_FILE_NOOSS, nil
		}
		signedURL, err := fileConf.Minio.PreSignedURL(req.Ext, dir)
		if err != nil {
			return errs.IMErrorCode_APP_FILE_SIGNERR, nil
		}
		return errs.IMErrorCode_SUCCESS, &apimodels.QryFileCredResp{
			OssType: apimodels.OssType_Minio,
			PreSignResp: &apimodels.PreSignResp{
				Url: signedURL,
			},
		}
	case fileengines.ChannelAws:
		if fileConf.S3 == nil {
			return errs.IMErrorCode_APP_FILE_NOOSS, nil
		}

		signedURL, err := fileConf.S3.PreSignedURL(req.Ext, dir)
		if err != nil {
			return errs.IMErrorCode_APP_FILE_SIGNERR, nil
		}
		return errs.IMErrorCode_SUCCESS, &apimodels.QryFileCredResp{
			OssType: apimodels.OssType_S3,
			PreSignResp: &apimodels.PreSignResp{
				Url: signedURL,
			},
		}
	case fileengines.ChannelOss:
		if fileConf.Oss == nil {
			return errs.IMErrorCode_APP_FILE_NOOSS, nil
		}
		signedURL, err := fileConf.Oss.PreSignedURL(req.Ext, dir)
		if err != nil {
			return errs.IMErrorCode_APP_FILE_SIGNERR, nil
		}
		resp := fileConf.Oss.PostSign(req.Ext, dir)
		return errs.IMErrorCode_SUCCESS, &apimodels.QryFileCredResp{
			OssType: apimodels.OssType_Oss,
			PreSignResp: &apimodels.PreSignResp{
				Url:         signedURL,
				ObjKey:      resp.ObjKey,
				Policy:      resp.Policy,
				SignVersion: resp.SignVersion,
				Credential:  resp.Credential,
				Date:        resp.Date,
				Signature:   resp.Signature,
			},
		}
	default:
		return errs.IMErrorCode_APP_FILE_NOOSS, nil
	}
}

type FileConfItem struct {
	AppKey     string `json:"app_key,omitempty"`
	FileEngine string `json:"file_engine"`
	//QiniuConfig *QiniuFileConfig `json:"qiniu,omitempty"`

	QiNiu *fileengines.QiNiuStorage
	Oss   *fileengines.OssStorage
	Minio *fileengines.MinioStorage
	S3    *fileengines.S3Storage
}

var fileConfCache *caches.LruCache
var fileLock *sync.RWMutex
var notExistFileConf *FileConfItem

func init() {
	fileConfCache = caches.NewLruCacheWithAddReadTimeout("fileconf_caches", 1000, nil, 5*time.Minute, 10*time.Minute)
	fileLock = &sync.RWMutex{}
	notExistFileConf = &FileConfItem{}
}

func GetFileConf(ctx context.Context, appkey string) *FileConfItem {
	if obj, exist := fileConfCache.Get(appkey); exist {
		return obj.(*FileConfItem)
	} else {
		fileLock.Lock()
		defer fileLock.Unlock()

		if obj, exist := fileConfCache.Get(appkey); exist {
			return obj.(*FileConfItem)
		} else { //load from db
			fileConf, err := loadFileConfFromDb(appkey)
			if err != nil {
				fileConf = notExistFileConf
			}
			fileConfCache.Add(appkey, fileConf)
			return fileConf
		}
	}
}

func loadFileConfFromDb(appkey string) (*FileConfItem, error) {
	dao := dbs.FileConfDao{}
	conf, err := dao.FindEnableFileConf(appkey)
	if err != nil {
		return nil, err
	}
	fileConf := &FileConfItem{
		AppKey:     appkey,
		FileEngine: conf.Channel,
	}
	var confData = make(map[string]interface{})
	_ = json.Unmarshal([]byte(conf.Conf), &confData)

	switch conf.Channel {
	case fileengines.ChannelQiNiu:
		c := utils.MapToStruct[fileengines.QiNiuConfig](confData)
		fileConf.QiNiu = fileengines.NewQiNiu(c)
	case fileengines.ChannelMinio:
		c := utils.MapToStruct[fileengines.MinioConfig](confData)
		fileConf.Minio = fileengines.NewMinio(c)
	case fileengines.ChannelOss:
		c := utils.MapToStruct[fileengines.OssConfig](confData)
		fileConf.Oss = fileengines.NewOss(c)
	case fileengines.ChannelAws:
		c := utils.MapToStruct[fileengines.S3Config](confData)
		fileConf.S3 = fileengines.NewS3Storage(fileengines.WithConf(c))
	}
	return fileConf, nil
}

func fileTypeToDir(fileType apimodels.FileType) string {
	switch fileType {
	case apimodels.FileType_Image:
		return "images"
	case apimodels.FileType_Video:
		return "videos"
	case apimodels.FileType_Audio:
		return "audios"
	case apimodels.FileType_File:
		return "files"
	case apimodels.FileType_Log:
		return "logs"
	default:
		return "files"
	}
}
