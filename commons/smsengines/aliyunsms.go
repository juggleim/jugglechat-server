package smsengines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/dara"
)

const (
	defaultAliyunSmsEndpoint        = "dysmsapi.aliyuncs.com"
	defaultAliyunSmsRegion          = "cn-hangzhou"
	defaultAliyunSmsTemplateCodeKey = "code"
	aliyunSmsConnectTimeout         = 5 * time.Second
	aliyunSmsReadTimeout            = 10 * time.Second
)

type aliyunSmsClient interface {
	SendSmsWithContext(context.Context, *dysms.SendSmsRequest, *dara.RuntimeOptions) (*dysms.SendSmsResponse, error)
}

// SMSRejectedError means a provider business response explicitly refused delivery.
// Transport errors and missing business status must never be classified by text.
type SMSRejectedError struct{ Code, RequestID string }

func (e *SMSRejectedError) Error() string {
	return fmt.Sprintf("aliyun sms rejected request: code=%s request_id=%s", e.Code, e.RequestID)
}

// AliyunSmsEngine sends template SMS messages through Alibaba Cloud SMS.
// One engine instance is created and cached for each app key.
type AliyunSmsEngine struct {
	clientMu sync.Mutex      `json:"-"`
	client   aliyunSmsClient `json:"-"`

	AccessKeyId     string `json:"access_key"`
	AccessKeySecret string `json:"access_secret"`
	RegionId        string `json:"region_id,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
	TemplateCodeKey string `json:"template_code_key,omitempty"`
}

func (eng *AliyunSmsEngine) Valid() bool {
	return eng != nil &&
		strings.TrimSpace(eng.AccessKeyId) != "" &&
		strings.TrimSpace(eng.AccessKeySecret) != "" &&
		strings.TrimSpace(eng.SignName) != "" &&
		strings.TrimSpace(eng.TemplateCode) != ""
}

// Prepare validates local configuration/client creation without sending a request.
func (eng *AliyunSmsEngine) Prepare() error {
	_, err := eng.createClient()
	return err
}

func (eng *AliyunSmsEngine) createClient() (aliyunSmsClient, error) {
	eng.clientMu.Lock()
	defer eng.clientMu.Unlock()

	if eng.client != nil {
		return eng.client, nil
	}
	if !eng.Valid() {
		return nil, fmt.Errorf("aliyun sms config invalid")
	}

	regionId := strings.TrimSpace(eng.RegionId)
	if regionId == "" {
		regionId = defaultAliyunSmsRegion
	}
	endpoint := strings.TrimSpace(eng.Endpoint)
	if endpoint == "" {
		endpoint = defaultAliyunSmsEndpoint
	}

	client, err := dysms.NewClient(&openapi.Config{
		AccessKeyId:     dara.String(eng.AccessKeyId),
		AccessKeySecret: dara.String(eng.AccessKeySecret),
		RegionId:        dara.String(regionId),
		Endpoint:        dara.String(endpoint),
		Protocol:        dara.String("HTTPS"),
		ConnectTimeout:  dara.Int(int(aliyunSmsConnectTimeout.Milliseconds())),
		ReadTimeout:     dara.Int(int(aliyunSmsReadTimeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("create aliyun sms client: %w", err)
	}
	eng.client = client
	return eng.client, nil
}

func (eng *AliyunSmsEngine) SmsSend(phone string, params map[string]interface{}) error {
	if strings.TrimSpace(phone) == "" || len(params) == 0 {
		return fmt.Errorf("aliyun sms params invalid")
	}

	client, err := eng.createClient()
	if err != nil {
		return err
	}

	templateParams := make(map[string]interface{}, len(params))
	for key, value := range params {
		templateParams[key] = value
	}
	templateCodeKey := strings.TrimSpace(eng.TemplateCodeKey)
	if templateCodeKey == "" {
		templateCodeKey = defaultAliyunSmsTemplateCodeKey
	}
	if templateCodeKey != defaultAliyunSmsTemplateCodeKey {
		if code, ok := templateParams[defaultAliyunSmsTemplateCodeKey]; ok {
			delete(templateParams, defaultAliyunSmsTemplateCodeKey)
			templateParams[templateCodeKey] = code
		}
	}
	templateParamJSON, err := json.Marshal(templateParams)
	if err != nil {
		return fmt.Errorf("marshal aliyun sms template params: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), aliyunSmsConnectTimeout+aliyunSmsReadTimeout)
	defer cancel()
	response, err := client.SendSmsWithContext(ctx, (&dysms.SendSmsRequest{}).
		SetPhoneNumbers(phone).
		SetSignName(eng.SignName).
		SetTemplateCode(eng.TemplateCode).
		SetTemplateParam(string(templateParamJSON)), &dara.RuntimeOptions{
		Autoretry:      dara.Bool(false),
		ConnectTimeout: dara.Int(int(aliyunSmsConnectTimeout.Milliseconds())),
		ReadTimeout:    dara.Int(int(aliyunSmsReadTimeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("aliyun sms delivery failed: %w", err)
	}
	if response == nil || response.Body == nil {
		return fmt.Errorf("aliyun sms delivery failed: empty response")
	}
	if code := dara.StringValue(response.Body.Code); code != "OK" {
		if code == "" {
			return fmt.Errorf("aliyun sms delivery failed: missing business status")
		}
		return &SMSRejectedError{Code: code, RequestID: dara.StringValue(response.Body.RequestId)}
	}
	return nil
}
