package services

import (
	"testing"

	"github.com/juggleim/jugglechat-server/commons/smsengines"
)

func TestSmsEngineConfAliyun(t *testing.T) {
	aliyun := &smsengines.AliyunSmsEngine{
		AccessKeyId:     "access-key-id",
		AccessKeySecret: "access-key-secret",
		SignName:        "JuggleChat",
		TemplateCode:    "SMS_123456",
	}
	conf := &SmsEngineConf{Channel: "aliyun", AliyunSmsEngine: aliyun}

	if got := conf.Engine(); got != aliyun {
		t.Fatalf("Engine() = %#v, want aliyun engine", got)
	}
}

func TestSmsEngineConfRejectsInvalidAliyun(t *testing.T) {
	conf := &SmsEngineConf{
		Channel: "aliyun",
		AliyunSmsEngine: &smsengines.AliyunSmsEngine{
			AccessKeyId:     "access-key-id",
			AccessKeySecret: "access-key-secret",
			TemplateCode:    "SMS_123456",
		},
	}

	if got := conf.Engine(); got != nil {
		t.Fatalf("Engine() = %#v, want nil", got)
	}
}
