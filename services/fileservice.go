package services

import (
	"appserver/configures"
	"context"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type UploadToken struct {
	Token  string `json:"token"`
	Domain string `json:"domain"`
}

func GetFileCred(ctx context.Context) *UploadToken {
	putPolicy := storage.PutPolicy{
		Scope: configures.Config.Qiniu.Bucket,
	}
	mac := qbox.NewMac(configures.Config.Qiniu.AccessKey, configures.Config.Qiniu.SecretKey)
	upToken := &UploadToken{
		Token:  putPolicy.UploadToken(mac),
		Domain: configures.Config.Qiniu.Domain,
	}
	return upToken
}
