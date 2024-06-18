package app

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"past-papers-web/internal/cache"
	"past-papers-web/internal/config"
	"past-papers-web/internal/helper"
	"past-papers-web/internal/mail"
)

var (
	port = flag.Int("port", 3000, "The server port")
)

type App struct {
	helper    *helper.Helper
	usercache *cache.Cache[string, interface{}]
	filecache *cache.Cache[string, []byte]
	mailer    *mail.Mail
}

func StartServer() {
	// Start the server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	app := NewApp()

	err = http.Serve(lis, app.Routes())
	if err != nil {
		return
	}
}

func NewApp() *App {
	config := config.NewConfig()
	usercache := cache.New[string, interface{}]()
	filecache := cache.New[string, []byte]()
	mailer := mail.New(config.SMTPFrom, config.SMTPPass, config.SMTPHost, config.SMTPPort)
	return &App{
		helper:    helper.NewHelper(config),
		usercache: usercache,
		filecache: filecache,
		mailer:    mailer,
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.Login)
	a.RegisterAdminRoutes("/admin", mux)
	mux.HandleFunc("/refresh-tree", a.RefreshTree)
	mux.HandleFunc("/register", a.Register)
	mux.HandleFunc("/content/", a.loginProtect(a.ContentHandler))
	return mux
}

func (a *App) RefreshTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	helper.RefreshTree(config.NewConfig(), a.helper)
	w.WriteHeader(http.StatusOK)
	return
}

// Auth middleware
func (a *App) loginProtect(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("email")
		if err != nil { // No cookie
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if _, ok := a.usercache.Get(cookie.Value); ok { // Has user in cache
			next.ServeHTTP(w, r)
			return
		}
		if a.helper.CheckUser(cookie.Value) { // Has user in DB
			a.usercache.Set(cookie.Value, true, time.Duration(time.Hour*720))
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	})
}
