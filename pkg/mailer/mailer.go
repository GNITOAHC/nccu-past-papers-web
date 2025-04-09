package mailer

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
)

const (
	SyncMailerIndicator = -1
)

type mailContent struct {
	To      string
	Subject string
	Content string
}

type Mailer struct {
	from        string
	pass        string
	host        string
	port        string
	mailChan    chan mailContent
	synchronous bool
}

// Create a new mailer, set queueSize to SyncMailerIndicator to send mail synchronously
func New(from, pass, host, port string, queueSize int) *Mailer {
	if queueSize == SyncMailerIndicator {
		return &Mailer{from: from, pass: pass, host: host, port: port, synchronous: true}
	}

	m := &Mailer{
		from:        from,
		pass:        pass,
		host:        host,
		port:        port,
		mailChan:    make(chan mailContent, queueSize),
		synchronous: false,
	}

	go m.mailWorker()

	return m
}

func (m *Mailer) mailWorker() {
	for mail := range m.mailChan {
		// Process each mail in the queue
		if err := m.send(mail.To, mail.Subject, mail.Content); err != nil {
			log.Println(err)
		} else {
			// Send success
		}
	}
}

// Send a mail, if the mailer is synchronous, the function will block until the mail is sent, otherwise it will return nil immediately
func (m *Mailer) Send(to, subject, content string) error {
	if m.synchronous {
		return m.send(to, subject, content)
	}

	mail := mailContent{To: to, Subject: subject, Content: content}
	m.mailChan <- mail
	return nil
}

func (m *Mailer) send(to, subject, content string) error {
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
