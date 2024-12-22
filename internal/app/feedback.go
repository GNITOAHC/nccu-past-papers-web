package app

import (
	"net/http"
	"strings"
)

func (a *App) Feedback(w http.ResponseWriter, r *http.Request) {
	feedback := r.FormValue("feedback")
	for _, m := range strings.Split(a.config.ADMIN_MAIL, ",") {
		err := a.mailer.Send(m, "Feedback", feedback)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Write([]byte("Feedback sent"))
}
