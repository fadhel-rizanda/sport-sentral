package mailer

import (
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Mailer struct {
	cfg  Config
	auth smtp.Auth
}

func New(cfg Config) *Mailer {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &Mailer{
		cfg:  cfg,
		auth: auth,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf(
		"%s:%d",
		m.cfg.Host, m.cfg.Port,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, body,
	))
	return smtp.SendMail(addr, m.auth, m.cfg.From, []string{to}, msg)
}
