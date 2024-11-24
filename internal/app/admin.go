package app

import (
	"encoding/json"
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

func (a *App) RegisterAdminRoutes(prefix string, mux *http.ServeMux) {
	mux.HandleFunc(prefix+"/", a.adminProtect(a.loginProtect(a.Admin)))
	// mux.HandleFunc(prefix+"/", a.loginProtect(a.Admin)) // For testing process
	// mux.HandleFunc(prefix+"/", a.Admin) // For testing process
	mux.HandleFunc(prefix+"/approve", a.adminProtect(a.loginProtect(a.ApproveRegistration)))
	// mux.HandleFunc(prefix+"/approve", a.ApproveRegistration) // For testing process
	mux.HandleFunc(prefix+"/delete", a.adminProtect(a.loginProtect(a.DeleteRegistration)))
	mux.HandleFunc(prefix+"/pr_approve", a.adminProtect(a.loginProtect(a.ApprovePullRequest)))
	mux.HandleFunc(prefix+"/pr_delete", a.adminProtect(a.loginProtect(a.DeletePullRequest)))
	return
}

func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"WaitingList": a.helper.GetWaitingList(),
		"PRList":      a.helper.GetPRList(a.config.GitHubAccessToken),
	}
	templates.Render(w, "admin.html", data)
	// a.tmplExecute(w, []string{"templates/admin.html"}, map[string]interface{}{
	// 	"WaitingList": a.helper.GetWaitingList(),
	// })
}

// ApproveRegistration approves the registration of the user.
//
// JSON body:
// - email: user's mail
// - name: user's name
// - studentId: user's student ID
//
// @return 200: Success
// @return 400: Bad request
// @return 405: Method not allowed
// @return 500: Internal server error (_e.g._ database error)
func (a *App) ApproveRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = a.helper.ApproveRegistration(data["email"], data["name"], data["studentId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (a *App) DeleteRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = a.helper.DeleteRegistration(data["email"], data["name"], data["studentId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (a *App) ApprovePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data map[string]int
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = a.helper.ApprovePullRequest(data["pr"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (a *App) DeletePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data map[string]int
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = a.helper.DeletePullRequest(data["pr"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}