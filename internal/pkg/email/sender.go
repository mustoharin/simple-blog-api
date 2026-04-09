package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// Sender sends emails via SMTP.
type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSender creates a new Sender with SMTP credentials.
func NewSender(host string, port int, username, password, from string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Message represents an outgoing email.
type Message struct {
	To      string
	Subject string
	Body    string // HTML body
}

// Send delivers the message via SMTP.
func (s *Sender) Send(msg Message) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/html", msg.Body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email to %s: %w", msg.To, err)
	}
	return nil
}
