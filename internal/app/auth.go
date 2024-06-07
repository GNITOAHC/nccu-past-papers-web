package app

import (
	"html/template"
	"net/http"
	"time"
)

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	// Register user
	email := r.FormValue("email")
	name := r.FormValue("name")
	studentId := r.FormValue("studentId")
	for _, v := range []string{email, name, studentId} {
		if v == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
	}
	success := a.helper.RegisterUser(email, name, studentId)
	if !success {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Success, please wait for approval"))
	return
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	renderTmpl := func() {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/entry.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if r.Method == "POST" {
		email := r.FormValue("email")
		http.SetCookie(w, &http.Cookie{ // Set a cookie
			Name:     "email",
			Value:    email,
			MaxAge:   3600,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		if _, ok := a.usercache.Get(email); ok { // Has user in cache
			http.Redirect(w, r, "/content", http.StatusSeeOther)
			return
		}
		if a.helper.CheckUser(email) { // Has user in DB
			a.usercache.Set(email, true, time.Duration(time.Hour*720)) // Set cache for 30 days
			http.Redirect(w, r, "/content", http.StatusSeeOther)
			return
		}

		renderTmpl()
		return
	}
	renderTmpl()
	return
}
