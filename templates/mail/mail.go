package mail

import (
	"bytes"
	"html/template"
	"past-papers-web/mailer"
)

func SendMail(m *mailer.Mailer, data map[string]interface{}, title string, tmpl, to []string) error {
	t, err := template.ParseFiles(tmpl...)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	for _, tt := range to {
		if err := m.Send(tt, title, buf.String()); err != nil {
			return err
		}
	}
	return nil
}
