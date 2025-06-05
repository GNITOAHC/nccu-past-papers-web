package app

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"past-papers-web/internal/database"
	"past-papers-web/templates"
	"past-papers-web/templates/mail"
	"strings"
	"time"
)

type UserReg struct {
	Email     string
	Name      string
	StudentId string
	OTP       string
}

func generateOTP() (string, error) {
	const otpLength = 6
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	charsetLen := len(charset)

	otp := make([]byte, otpLength)
	for i := 0; i < otpLength; i++ {
		// Generate a random index to pick a character from charset
		num, err := rand.Int(rand.Reader, big.NewInt(int64(charsetLen)))
		if err != nil {
			return "", err
		}
		otp[i] = charset[num.Int64()]
	}

	return string(otp), nil
}

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

	otp, err := generateOTP()
	if err != nil {
		http.Error(w, "Failed to generate OTP", http.StatusInternalServerError)
		return
	}
	a.otpcache.Set(email, UserReg{
		Email: email, Name: name, StudentId: studentId, OTP: otp,
	}, time.Duration(time.Minute*10))

	w.Write([]byte("Success, please check your email and for one time passcode.")) // Write to response first

	// Send mail to user
	data := map[string]interface{}{"Name": name, "OTP": otp}
	mail.SendMail(a.mailer, data, "OTP Verification", []string{"templates/mail/otp.html"}, []string{email})
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

		log.Printf("Login attempt for email: %s", email)

		if _, ok := a.usercache.Get(email); ok { // Has user in cache
			http.Redirect(w, r, "/content", http.StatusSeeOther)
			return
		}
		res, err := database.CheckUser(email, r.Context())
		if err != nil {
			log.Print("Error checking user in login: ", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if res.Exist {
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

func (a *App) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	email := r.FormValue("email")
	otp := r.FormValue("otp")
	if email == "" || otp == "" {
		log.Println("Missing Email or OTP; user: ", email, "; otp: ", otp)
		http.Error(w, "Missing Email or OTP", http.StatusBadRequest)
		return
	}
	info, ok := a.otpcache.Get(email)
	if !ok {
		http.Error(w, "OTP expired or not found", http.StatusBadRequest)
		return
	}
	if info.OTP != otp {
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	err := database.RegisterUser(info.Email, info.Name, info.StudentId, r.Context())
	if err != nil {
		log.Println("Error registering user: (", info.Email, info.Name, info.StudentId, ") ", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success, you can now login.")) // Write to response first
	a.otpcache.Delete(email)

	data := map[string]interface{}{"Name": info.Name, "Email": email, "StudentId": info.StudentId}
	err = mail.SendMail(a.mailer, data, "New Registration", []string{"templates/mail/regadminnotify.html"}, strings.Split(a.config.ADMIN_MAIL, ","))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = mail.SendMail(a.mailer, map[string]interface{}{"Name": info.Name}, "OTP Verified", []string{"templates/mail/register.html"}, []string{email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
