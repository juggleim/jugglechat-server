package emailengines

var (
	DefaultEmailEngine IEmailEngine = &NilEmailEngine{}
)

type IEmailEngine interface {
	SendMail(toAddress string, subject string, txtBody, htmlBody string) error
}

type NilEmailEngine struct{}

func (engine *NilEmailEngine) SendMail(toAddress string, subject string, txtBody, htmlBody string) error {
	return nil
}
