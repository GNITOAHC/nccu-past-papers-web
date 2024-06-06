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
)

var (
	port = flag.Int("port", 3000, "The server port")
)

type App struct {
	helper    *helper.Helper
	usercache *cache.Cache[string, interface{}]
	filecache *cache.Cache[string, []byte]
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
	return &App{
		helper:    helper.NewHelper(config),
		usercache: usercache,
		filecache: filecache,
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.Login)
	mux.HandleFunc("/register", a.Register)
	mux.HandleFunc("/content/", a.loginProtect(a.ContentHandler))
	return mux
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
