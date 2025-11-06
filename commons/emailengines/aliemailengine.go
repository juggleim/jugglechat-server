package emailengines

import (
	"encoding/json"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dm20151123 "github.com/alibabacloud-go/dm-20151123/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type AliEmailEngine struct {
	client          *dm20151123.Client `json:"-"`
	AccessKeyId     string             `json:"access_key"`
	AccessKeySecret string             `json:"access_secret"`
	FromEmail       string             `json:"from_email"`
	FromAlias       string             `json:"from_alias"`
}

func (engine *AliEmailEngine) createClient() (*dm20151123.Client, error) {
	if engine.client != nil {
		return engine.client, nil
	}
	akConfig := new(credential.Config).SetType("access_key").
		SetAccessKeyId(engine.AccessKeyId).
		SetAccessKeySecret(engine.AccessKeySecret)
	credential, err := credential.NewCredential(akConfig)
	if err != nil {
		return nil, err
	}
	config := &openapi.Config{
		Credential: credential,
	}
	config.Endpoint = tea.String("dm.aliyuncs.com")
	result, err := dm20151123.NewClient(config)
	if err == nil {
		engine.client = result
	}
	return result, err
}

func (engine *AliEmailEngine) SendMail(toAddress string, subject string, txtBody, htmlBody string) error {
	client, err := engine.createClient()
	if err != nil {
		return err
	}
	singleSendMailRequest := &dm20151123.SingleSendMailRequest{
		AccountName:    tea.String(engine.FromEmail),
		FromAlias:      tea.String(engine.FromAlias),
		AddressType:    tea.Int32(1),
		ReplyToAddress: tea.Bool(false),
		ToAddress:      tea.String(toAddress),
		Subject:        tea.String(subject),
		TextBody:       tea.String(txtBody),
		HtmlBody:       tea.String(htmlBody),
	}
	err = func() (e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				e = r
			}
		}()
		resp, err := client.SingleSendMailWithOptions(singleSendMailRequest, &util.RuntimeOptions{})
		if err != nil {
			return err
		}
		fmt.Println(util.ToJSONString(resp))
		return nil
	}()
	if err != nil {
		var error = &tea.SDKError{}
		if t, ok := err.(*tea.SDKError); ok {
			error = t
		} else {
			error.Message = tea.String(err.Error())
		}
		fmt.Println(tea.StringValue(error.Message))
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			fmt.Println(m["Recommend"])
		}
		_, err = util.AssertAsString(error.Message)
		if err != nil {
			return err
		}
	}
	return err
}
