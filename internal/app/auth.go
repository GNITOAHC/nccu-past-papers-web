package app

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"past-papers-web/templates"
	"strings"
	"time"
)

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	// Register user
	email := r.FormValue("email")
	http.SetCookie(w, &http.Cookie{ // Set a cookie
		Name:     "email",
		Value:    email,
		MaxAge:   3600,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

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

	// w.Write([]byte("Success, please check your email and wait for approval.")) // Write to response first

	otp := generateOTP()
	a.otpCache.Set(email, otp, 5*time.Minute)

	// Send mail to user
	t, err := template.ParseFiles("templates/mail/otp.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf := new(bytes.Buffer)
	// data := map[string]interface{}{"Name": name}
	data := map[string]interface{}{"Name": name, "OTP": otp}

	if err = t.Execute(buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.mailer.Send(email, "Registration", buf.String())
	w.Write([]byte("Success, please check your email for the OTP to complete verification."))

	// Send mail to administator
	t, err = template.ParseFiles("templates/mail/regadminnotify.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf = new(bytes.Buffer)
	data = map[string]interface{}{"Name": name, "Email": email, "StudentId": studentId}
	if err = t.Execute(buf, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, admin := range strings.Split(a.config.ADMIN_MAIL, ",") {
		a.mailer.Send(strings.TrimSpace(admin), "New Registration", buf.String())
	}

	return
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
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

		templates.Render(w, "entry.html", nil)
		return
	}
	templates.Render(w, "entry.html", nil)
	return
}

func generateOTP() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (a *App) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Retrieve the "email" cookie
		cookie, err := r.Cookie("email")
		if err != nil {
			http.Error(w, "Email not found", http.StatusBadRequest)
			return
		}
		email := cookie.Value
		otpInput := r.FormValue("otp")

		// Retrieve the stored OTP from the cache
		storedOTP, ok := a.otpCache.Get(email)
		if !ok || storedOTP != otpInput {
			http.Error(w, "Invalid or expired OTP", http.StatusUnauthorized)
			return
		}

		// clear the OTP and redirect
		a.otpCache.Delete(email)
		http.Redirect(w, r, "/content", http.StatusSeeOther)
		return
	}
	templates.Render(w, "otp_verify.html", nil)
}
