package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/ReadingGarden/back-go/internal/config"
)

type Mailer interface {
	Send(email, title, content string) error
}

type SMTPMailer struct {
	account  string
	password string
}

func NewSMTPMailer(cfg config.EmailConfig) *SMTPMailer {
	return &SMTPMailer{
		account:  cfg.Account,
		password: cfg.Password,
	}
}

func (m *SMTPMailer) Send(email, title, content string) error {
	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", &tls.Config{ServerName: "smtp.gmail.com"})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return err
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", m.account, m.password, "smtp.gmail.com")); err != nil {
		return err
	}
	if err := client.Mail(m.account); err != nil {
		return err
	}
	if err := client.Rcpt(email); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\n\r\n%s", title, m.account, email, content)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return err
	}

	return writer.Close()
}
