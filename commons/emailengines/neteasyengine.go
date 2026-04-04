package emailengines

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	gomail "github.com/go-mail/mail/v2"
)

type NeteasyEmailEngine struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	AuthCode           string `json:"auth_code"`
	Password           string `json:"password,omitempty"`
	FromEmail          string `json:"from_email"`
	FromAlias          string `json:"from_alias"`
	UseStartTLS        bool   `json:"use_starttls"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
}

func (engine *NeteasyEmailEngine) SendMail(toAddress string, subject string, txtBody, htmlBody string) error {
	host := strings.TrimSpace(engine.Host)
	if host == "" {
		host = "smtp.163.com"
	}
	port := engine.Port
	if port <= 0 {
		port = 465
	}
	fromEmail := strings.TrimSpace(engine.FromEmail)
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(engine.Username)
	}
	authSecret := strings.TrimSpace(engine.AuthCode)
	if authSecret == "" {
		authSecret = strings.TrimSpace(engine.Password)
	}
	if fromEmail == "" || strings.TrimSpace(engine.Username) == "" || authSecret == "" || strings.TrimSpace(toAddress) == "" {
		return fmt.Errorf("neteasy mail config invalid")
	}
	timeout := 10 * time.Second
	if engine.TimeoutSeconds > 0 {
		timeout = time.Duration(engine.TimeoutSeconds) * time.Second
	}
	return engine.sendByGoMail(host, port, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody, timeout)
}

func (engine *NeteasyEmailEngine) sendByGoMail(host string, port int, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody string, timeout time.Duration) error {
	m := gomail.NewMessage()
	if strings.TrimSpace(engine.FromAlias) != "" {
		m.SetAddressHeader("From", fromEmail, strings.TrimSpace(engine.FromAlias))
	} else {
		m.SetHeader("From", fromEmail)
	}
	m.SetHeader("To", strings.TrimSpace(toAddress))
	m.SetHeader("Subject", strings.TrimSpace(subject))

	txtBody = strings.TrimSpace(txtBody)
	htmlBody = strings.TrimSpace(htmlBody)
	if txtBody == "" && htmlBody == "" {
		txtBody = ""
	}
	if txtBody != "" {
		m.SetBody("text/plain; charset=UTF-8", txtBody)
		if htmlBody != "" {
			m.AddAlternative("text/html; charset=UTF-8", htmlBody)
		}
	} else {
		m.SetBody("text/html; charset=UTF-8", htmlBody)
	}

	dialer := gomail.NewDialer(host, port, strings.TrimSpace(engine.Username), authSecret)
	dialer.Timeout = timeout
	dialer.TLSConfig = &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: engine.InsecureSkipVerify,
	}

	if engine.UseStartTLS {
		dialer.SSL = false
		dialer.StartTLSPolicy = gomail.MandatoryStartTLS
	} else {
		dialer.SSL = true
		dialer.StartTLSPolicy = gomail.NoStartTLS
	}
	return dialer.DialAndSend(m)
}
