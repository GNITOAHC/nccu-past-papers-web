package app

import (
	"net/http"
)

func (a *App) adminProtect(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("email")
		if err != nil { // No cookie
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if a.helper.IsAdmin(cookie.Value) { // Has admin user in DB
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	})
}

func (a *App) RegisterAdminRoutes(prefix string, mux *http.ServeMux) {
	mux.HandleFunc(prefix+"/", a.adminProtect(a.loginProtect(a.Admin)))
	// mux.HandleFunc(prefix+"/", a.loginProtect(a.Admin)) // For testing process
	return
}

func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	a.tmplExecute(w, []string{"templates/admin.html"}, map[string]interface{}{
		"WaitingList": a.helper.GetWaitingList(),
	})
	return
}
