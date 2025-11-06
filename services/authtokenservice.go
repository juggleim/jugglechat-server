package services

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/juggleim/jugglechat-server/commons/appinfos"
	utils "github.com/juggleim/jugglechat-server/commons/tools"
	"github.com/juggleim/jugglechat-server/services/pbobjs"
)

type ImToken struct {
	AppKey    string
	UserId    string
	DeviceId  string
	TokenTime int64
}

func (t ImToken) ToTokenString(secureKey []byte) (string, error) {
	tokenValue := &pbobjs.AuthTokenValue{
		UserId:    t.UserId,
		DeviceId:  t.DeviceId,
		TokenTime: t.TokenTime,
	}
	tokenBs, err := utils.PbMarshal(tokenValue)
	if err == nil {
		encryptToken, err := encrypt(tokenBs, secureKey)
		if err == nil {
			tokenWrap := &pbobjs.AuthToken{
				Appkey:     t.AppKey,
				TokenValue: encryptToken,
			}
			tokenWrapBs, err := utils.PbMarshal(tokenWrap)
			if err == nil {
				bas64TokenStr := base64.URLEncoding.EncodeToString(tokenWrapBs)
				return bas64TokenStr, nil
			} else {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return "", err
}

func GenerateToken(appkey, userId string) string {
	token := ""
	t := &ImToken{
		AppKey:    appkey,
		UserId:    userId,
		TokenTime: time.Now().UnixMilli(),
	}
	if appInfo, exist := appinfos.GetAppInfo(appkey); exist {
		token, _ = t.ToTokenString([]byte(appInfo.AppSecureKey))
	}
	return token
}

func encrypt(dataBs, secureKeyBs []byte) ([]byte, error) {
	return utils.AesEncrypt(dataBs, secureKeyBs)
}
func decrypt(cryptedData, secureKeyBs []byte) ([]byte, error) {
	return utils.AesDecrypt(cryptedData, secureKeyBs)
}

func ParseTokenString(tokenStr string) (*pbobjs.AuthToken, error) {
	tokenWrap := &pbobjs.AuthToken{}
	tokenWrapBs, err := base64.URLEncoding.DecodeString(tokenStr)
	if err != nil {
		tokenWrapBs, err = base64.StdEncoding.DecodeString(tokenStr)
	}
	if err == nil {
		err = utils.PbUnMarshal(tokenWrapBs, tokenWrap)
	}
	return tokenWrap, err
}

func ParseToken(tokenWrap *pbobjs.AuthToken, secureKey []byte) (ImToken, error) {
	token := ImToken{
		AppKey: tokenWrap.Appkey,
	}
	cryptedToken := tokenWrap.TokenValue
	tokenBs, err := decrypt(cryptedToken, secureKey)
	if err == nil {
		tokenValue := &pbobjs.AuthTokenValue{}
		err = utils.PbUnMarshal(tokenBs, tokenValue)
		if err != nil {
			return token, err
		}
		if tokenValue.UserId == "" {
			return token, errors.New("invalid token")
		}
		token.UserId = tokenValue.UserId
		token.DeviceId = tokenValue.DeviceId
		token.TokenTime = tokenValue.TokenTime
	}
	return token, err
}
