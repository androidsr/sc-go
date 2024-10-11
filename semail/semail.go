package semail

import (
	"github.com/androidsr/sc-go/syaml"

	"gopkg.in/gomail.v2"
)

const (
	TEXT = "plain"
	HTML = "html"
)

var (
	host     string
	port     int
	username string
	password string
)

func New(config *syaml.EmailInfo) {
	host = config.Host
	port = config.Port
	username = config.Username
	password = config.Password
}

func SendEmail(from, to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	n := gomail.NewDialer(host, port, username, password)
	if err := n.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
