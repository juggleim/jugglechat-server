package emailengines

import (
	"encoding/base64"
	"fmt"

	"github.com/juggleim/jugglechat-server/commons/tools"
)

type EngagelabEmailEngine struct {
	Url       string `json:"url"`
	ApiUser   string `json:"api_user"`
	ApiKey    string `json:"api_key"`
	FromEmail string `json:"from_email"`
}

func (engine *EngagelabEmailEngine) SendMail(toAddress string, subject string, textBody, htmlBody string) error {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", engine.ApiUser, engine.ApiKey)))
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Basic " + auth,
	}
	body := &EngagelabEmailMsg{
		From: engine.FromEmail,
		To:   []string{toAddress},
		Body: &EngagelabEmailMsgBody{
			Subject: subject,
			Content: &EngagelabEmailMsgBodyContent{
				Html: htmlBody,
				Text: textBody,
			},
		},
	}
	resp, code, err := tools.HttpDo("POST", engine.Url, headers, tools.ToJson(body))
	if err != nil || code != 200 {
		fmt.Println("engagelab send email failed:", err, code, resp)
	}
	return err
}

type EngagelabEmailMsg struct {
	From string                 `json:"from"`
	To   []string               `json:"to"`
	Body *EngagelabEmailMsgBody `json:"body"`
}

type EngagelabEmailMsgBody struct {
	Subject string                        `json:"subject"`
	Content *EngagelabEmailMsgBodyContent `json:"content"`
}

type EngagelabEmailMsgBodyContent struct {
	Html        string `json:"html"`
	Text        string `json:"text"`
	PreviewText string `json:"preview_text"`
}
