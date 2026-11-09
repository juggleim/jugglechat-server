package smsengines

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/dara"
)

type fakeAliyunSmsClient struct {
	request *dysms.SendSmsRequest
	runtime *dara.RuntimeOptions
	resp    *dysms.SendSmsResponse
	err     error
}

func (client *fakeAliyunSmsClient) SendSmsWithContext(_ context.Context, request *dysms.SendSmsRequest, runtime *dara.RuntimeOptions) (*dysms.SendSmsResponse, error) {
	client.request = request
	client.runtime = runtime
	return client.resp, client.err
}

func newTestAliyunSmsEngine(client aliyunSmsClient) *AliyunSmsEngine {
	return &AliyunSmsEngine{
		client:          client,
		AccessKeyId:     "access-key-id",
		AccessKeySecret: "access-key-secret",
		SignName:        "JuggleChat",
		TemplateCode:    "SMS_123456",
	}
}

func TestAliyunSmsEngineSmsSend(t *testing.T) {
	client := &fakeAliyunSmsClient{resp: (&dysms.SendSmsResponse{}).SetBody(
		(&dysms.SendSmsResponseBody{}).SetCode("OK").SetRequestId("request-id"),
	)}
	engine := newTestAliyunSmsEngine(client)

	if err := engine.SmsSend("13800138000", map[string]interface{}{"code": "123456"}); err != nil {
		t.Fatalf("SmsSend returned error: %v", err)
	}
	if got := dara.StringValue(client.request.PhoneNumbers); got != "13800138000" {
		t.Fatalf("PhoneNumbers = %q", got)
	}
	if got := dara.StringValue(client.request.SignName); got != "JuggleChat" {
		t.Fatalf("SignName = %q", got)
	}
	if got := dara.StringValue(client.request.TemplateCode); got != "SMS_123456" {
		t.Fatalf("TemplateCode = %q", got)
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(dara.StringValue(client.request.TemplateParam)), &params); err != nil {
		t.Fatalf("invalid TemplateParam: %v", err)
	}
	if got := params["code"]; got != "123456" {
		t.Fatalf("TemplateParam code = %#v", got)
	}
	if client.runtime.Autoretry == nil || dara.BoolValue(client.runtime.Autoretry) {
		t.Fatal("Autoretry should be disabled")
	}
}

func TestAliyunSmsEngineUsesConfiguredTemplateCodeKey(t *testing.T) {
	client := &fakeAliyunSmsClient{resp: (&dysms.SendSmsResponse{}).SetBody(
		(&dysms.SendSmsResponseBody{}).SetCode("OK"),
	)}
	engine := newTestAliyunSmsEngine(client)
	engine.TemplateCodeKey = "verifyCode"

	if err := engine.SmsSend("13800138000", map[string]interface{}{"code": "654321"}); err != nil {
		t.Fatalf("SmsSend returned error: %v", err)
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(dara.StringValue(client.request.TemplateParam)), &params); err != nil {
		t.Fatalf("invalid TemplateParam: %v", err)
	}
	if params["verifyCode"] != "654321" {
		t.Fatalf("TemplateParam = %#v", params)
	}
	if _, exists := params["code"]; exists {
		t.Fatalf("default code key should be replaced: %#v", params)
	}
}

func TestAliyunSmsEngineRejectsProviderFailure(t *testing.T) {
	client := &fakeAliyunSmsClient{resp: (&dysms.SendSmsResponse{}).SetBody(
		(&dysms.SendSmsResponseBody{}).SetCode("isv.SMS_SIGNATURE_ILLEGAL").SetRequestId("request-id"),
	)}
	engine := newTestAliyunSmsEngine(client)

	err := engine.SmsSend("13800138000", map[string]interface{}{"code": "123456"})
	var rejected *SMSRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("expected typed provider rejection, got %T", err)
	}
	if err == nil || !strings.Contains(err.Error(), "isv.SMS_SIGNATURE_ILLEGAL") {
		t.Fatalf("expected provider rejection, got %v", err)
	}
	if strings.Contains(err.Error(), engine.AccessKeySecret) {
		t.Fatal("error leaked access key secret")
	}
}

func TestAliyunSmsEnginePropagatesClientError(t *testing.T) {
	client := &fakeAliyunSmsClient{err: errors.New("network error")}
	engine := newTestAliyunSmsEngine(client)

	err := engine.SmsSend("13800138000", map[string]interface{}{"code": "123456"})
	if err == nil || !strings.Contains(err.Error(), "network error") {
		t.Fatalf("expected client error, got %v", err)
	}
	var rejected *SMSRejectedError
	if errors.As(err, &rejected) {
		t.Fatal("network error classified as rejection")
	}
}

func TestAliyunSmsMissingStatusIsUnknown(t *testing.T) {
	client := &fakeAliyunSmsClient{resp: (&dysms.SendSmsResponse{}).SetBody(&dysms.SendSmsResponseBody{})}
	err := newTestAliyunSmsEngine(client).SmsSend("13800138000", map[string]interface{}{"code": "123456"})
	var rejected *SMSRejectedError
	if err == nil || errors.As(err, &rejected) {
		t.Fatalf("missing status must be unknown: %v", err)
	}
}

func TestAliyunSmsEngineValid(t *testing.T) {
	engine := newTestAliyunSmsEngine(&fakeAliyunSmsClient{})
	if !engine.Valid() {
		t.Fatal("complete config should be valid")
	}
	engine.SignName = ""
	if engine.Valid() {
		t.Fatal("config without sign_name should be invalid")
	}
}
