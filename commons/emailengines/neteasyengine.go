package emailengines

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
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
	if !engine.UseStartTLS {
		return engine.sendWithTLS(host, port, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody, timeout)
	}
	return engine.sendWithStartTLS(host, port, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody, timeout)
}

func (engine *NeteasyEmailEngine) sendWithTLS(host string, port int, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody string, timeout time.Duration) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: engine.InsecureSkipVerify,
	})
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	return engine.sendWithClient(client, host, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody)
}

func (engine *NeteasyEmailEngine) sendWithStartTLS(host string, port int, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody string, timeout time.Duration) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		err = client.StartTLS(&tls.Config{
			ServerName:         host,
			InsecureSkipVerify: engine.InsecureSkipVerify,
		})
		if err != nil {
			client.Close()
			return err
		}
	}
	defer client.Close()
	return engine.sendWithClient(client, host, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody)
}

func (engine *NeteasyEmailEngine) sendWithClient(client *smtp.Client, host, fromEmail, authSecret, toAddress, subject, txtBody, htmlBody string) error {
	auth := smtp.PlainAuth("", strings.TrimSpace(engine.Username), authSecret, host)
	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(fromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(strings.TrimSpace(toAddress)); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	msg := buildMimeMessage(fromEmail, strings.TrimSpace(toAddress), strings.TrimSpace(engine.FromAlias), subject, txtBody, htmlBody)
	_, err = w.Write([]byte(msg))
	if err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMimeMessage(fromEmail, toAddress, fromAlias, subject, txtBody, htmlBody string) string {
	fromHeader := fromEmail
	if fromAlias != "" {
		fromHeader = fmt.Sprintf("%s <%s>", fromAlias, fromEmail)
	}
	txtBody = strings.ReplaceAll(txtBody, "\n", "\r\n")
	htmlBody = strings.ReplaceAll(htmlBody, "\n", "\r\n")
	subject = strings.ReplaceAll(subject, "\r", "")
	subject = strings.ReplaceAll(subject, "\n", "")
	if htmlBody == "" {
		return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", fromHeader, toAddress, subject, txtBody)
	}
	boundary := "====_JuggleChatMailBoundary_163_===="
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n--%s--\r\n",
		fromHeader, toAddress, subject, boundary, boundary, txtBody, boundary, htmlBody, boundary)
}
