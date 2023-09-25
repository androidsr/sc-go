package semail

import (
	"fmt"
	"net/smtp"

	"github.com/androidsr/sc-go/syaml"
)

var (
	host     string
	port     string
	username string
	password string
)

func New(config syaml.EmailInfo) {
	host = config.Host
	port = config.Port
	username = config.Username
	password = config.Password
}

func SendEmail(to, title, body string) error {
	auth := smtp.PlainAuth("", username, password, host)
	message := "Subject: " + title + "\n" + body
	err := smtp.SendMail(host+":"+fmt.Sprint(port), auth, username, []string{to}, []byte(message))
	if err != nil {
		return err
	}
	return nil
}
