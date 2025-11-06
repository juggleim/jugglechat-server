package services

import (
	"encoding/base64"
	"time"

	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services/pbobjs"
)

func CheckApiKey(apiKey string, appkey, secureKey string) bool {
	bs, err := base64.URLEncoding.DecodeString(apiKey)
	if err != nil {
		return false
	}
	decodedBs, err := utils.AesDecrypt(bs, []byte(secureKey))
	if err != nil {
		return false
	}
	var apikey pbobjs.ApiKey
	err = utils.PbUnMarshal(decodedBs, &apikey)
	if err != nil {
		return false
	}
	if apikey.Appkey != appkey {
		return false
	}
	return true
}

func GenerateApiKey(appkey, secureKey string) (string, error) {
	apikey := &pbobjs.ApiKey{
		Appkey:      appkey,
		CreatedTime: time.Now().UnixMilli(),
	}
	bs, _ := utils.PbMarshal(apikey)
	encodedBs, err := utils.AesEncrypt(bs, []byte(secureKey))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encodedBs), nil
}
