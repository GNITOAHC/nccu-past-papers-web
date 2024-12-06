package app

import (
	"net/http"
	"past-papers-web/templates"
)

type QAItem struct {
	Question string
	Answer   string
}

func (a *App) QAPage(w http.ResponseWriter, r *http.Request) {
	qaList := []QAItem{
		{"What is this website?", "This website provides access to past papers for NCCU courses."},
		{"How can I download the papers?", "You can download the papers by clicking the download link provided."},
	}

	templates.Render(w, "faq.html", map[string]interface{}{
		"qaList": qaList,
	})
}
