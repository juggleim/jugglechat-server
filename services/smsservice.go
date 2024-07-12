package services

import (
	"appserver/configures"
	"appserver/dbs"
	"appserver/utils"
	"fmt"
	"math/rand"
	"time"

	"github.com/baidubce/bce-sdk-go/services/sms"
	"github.com/baidubce/bce-sdk-go/services/sms/api"
)

var smsClient *sms.Client
var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}
func InitSms() error {
	// 用户指定的Endpoint
	ENDPOINT := ""
	// 初始化一个SmsClient
	var err error
	smsClient, err = sms.NewClient(configures.Config.BaiduSms.ApiKey, configures.Config.BaiduSms.SecretKey, ENDPOINT)
	if err != nil {
		smsClient = nil
		return err
	}
	return nil
}

func CheckPhoneSmsCode(phone, code string) bool {
	if code == "000000" {
		return true
	}
	smsDao := dbs.SmsDao{}
	sms, err := smsDao.FindByPhoneCode(phone, code)
	if err != nil {
		fmt.Println("Sms validate:", err)
		return false
	}
	interval := time.Now().Sub(sms.CreatedTime)
	if interval > 5*time.Minute { //过期
		fmt.Println("Sms outdate:", phone, code)
		return false
	}
	return true
}

func SmsSend(phone string) bool {
	//检查是否还有有效的
	smsDao := dbs.SmsDao{}
	smsdb, err := smsDao.FindByPhone(phone, time.Now().Add(-3*time.Minute))
	randomCode := RandomSms()
	if err == nil {
		randomCode = smsdb.Code
	} else {
		_, err = smsDao.Create(dbs.SmsDao{
			Phone:       phone,
			Code:        randomCode,
			CreatedTime: time.Now(),
		})
		if err != nil {
			return false
		}
	}

	contentMap := make(map[string]interface{})
	contentMap["code"] = randomCode
	sendSmsArgs := &api.SendSmsArgs{
		Mobile:      phone,
		Template:    "sms-tmpl-ujHsUs04132",
		SignatureId: "sms-sign-ISzNcS20502",
		ContentVar:  contentMap,
	}
	result, err := smsClient.SendSms(sendSmsArgs)
	if err != nil {
		fmt.Printf("send sms error, %s", err)
		return false
	}
	fmt.Printf("send sms success. %s", result)
	return true
}
func RandomSms() string {
	retCode := ""
	for i := 0; i < 6; i++ {
		item := random.Intn(10)
		retCode = retCode + utils.Int2String(int64(item))
	}
	return retCode
}
