package mail

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
)

type Mail struct {
	from string
	pass string
	host string
	port string
}

func New(from string, pass string) *Mail {
	return &Mail{from: from, pass: pass, host: "smtp.gmail.com", port: "587"}
}

func (m *Mail) Send(to, subject, content string) error {
	// Authentication
	auth := smtp.PlainAuth("", m.from, m.pass, m.host)
    // Build mail
    var body bytes.Buffer
	mime := "\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
    body.Write([]byte(fmt.Sprintf("Subject: %s \n %s \n\n %s \n", subject, mime, content)))

	err := smtp.SendMail(m.host+":"+m.port, auth, m.from, []string{to}, body.Bytes())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
