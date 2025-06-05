package app

import (
	"net/http"
	"past-papers-web/templates"
	"strings"
)

func (a *App) adminProtect(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("email")
		if err != nil { // No cookie
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if ok := func() bool {
			for _, mail := range strings.Split(a.config.ADMIN_MAIL, ",") {
				if strings.TrimSpace(mail) == cookie.Value {
					return true
				}
			}
			return false
		}(); ok {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	})
}

func (a *App) AdminRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.adminProtect(a.loginProtect(a.Admin)))
	return mux
}

func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	templates.Render(w, "admin.html", nil)
}
