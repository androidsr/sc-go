package semail

import (
	"io"
	"net/smtp"
	"strings"

	"github.com/androidsr/sc-go/syaml"
)

const (
	TEXT = "plain"
	HTML = "html"
)

var (
	host     string
	port     string
	username string
	password string
)

func New(config *syaml.EmailInfo) {
	host = config.Host
	port = config.Port
	username = config.Username
	password = config.Password
}

func SendEmail(mailtype, to, subject, body string) error {
	auth := smtp.PlainAuth("", username, password, host)

	var contentType string
	if mailtype == HTML {
		contentType = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + username + "<" + username + ">" + "\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host+":"+port, auth, username, send_to, msg)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
