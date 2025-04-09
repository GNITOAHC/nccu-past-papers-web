package mailer

import (
	"os"
	"past-papers-web/pkg/dotenv"
	"testing"
	"time"
)

func TestAsyncSend(t *testing.T) {
	dotenv.Load("../.env")
	from := os.Getenv("SMTP_FROM")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	test_mail_dest := os.Getenv("TEST_MAIL_DEST")

	m := New(from, pass, host, port, 10)
	for i := 0; i < 2; i++ {
		m.Send(test_mail_dest, "Test subject", "<p style='color: blue'>Test body</p>")
	}

	t.Log("All mail queued")

	// Check if the mail was sent
	for {
		if len(m.mailChan) != 0 {
			t.Log("Still sending... ", len(m.mailChan), " mails left")
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}
}

func TestSyncSend(t *testing.T) {
	dotenv.Load("../.env")
	from := os.Getenv("SMTP_FROM")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	test_mail_dest := os.Getenv("TEST_MAIL_DEST")

	m := New(from, pass, host, port, SyncMailerIndicator)
	for i := 0; i < 2; i++ {
		err := m.Send(test_mail_dest, "Test subject", "<p style='color: blue'>Test body</p>")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("Mail ", i, " sent")
	}
}
