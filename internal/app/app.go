package app

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"past-papers-web/cache"
	"past-papers-web/internal/config"
	"past-papers-web/internal/helper"
	"past-papers-web/mailer"
	"past-papers-web/templates"

	"github.com/joho/godotenv"
)

var (
	port = flag.Int("port", 3000, "The server port")
)

type App struct {
	helper            *helper.Helper
	config            *config.Config
	usercache         *cache.Cache[string, interface{}] // Cache for users who are logged in
	filecache         *cache.Cache[string, []byte]      // Cache for files from the server
	chatcache         *cache.Cache[string, string]      // Cache for uploaded files to Gemini
	mailer            *mailer.Mailer
	GitHubAccessToken string
}

func StartServer() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file, using system environment variables")
	}

	// Start the server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	app := NewApp()

	templates.NewTemplates()

	err = http.Serve(lis, app.Routes())
	if err != nil {
		return
	}
}

func NewApp() *App {
	config := config.NewConfig()
	usercache := cache.New[string, interface{}]()
	filecache := cache.New[string, []byte]()
	chatcache := cache.New[string, string]()
	mailer := mailer.New(config.SMTPFrom, config.SMTPPass, config.SMTPHost, config.SMTPPort, 10)
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	return &App{
		helper:            helper.NewHelper(config),
		config:            config,
		usercache:         usercache,
		filecache:         filecache,
		chatcache:         chatcache,
		mailer:            mailer,
		GitHubAccessToken: token,
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.Login)
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	a.RegisterAdminRoutes("/admin", mux)
	mux.HandleFunc("/refresh-tree", a.RefreshTree)
	mux.HandleFunc("/register", a.Register)
	mux.HandleFunc("/content/", a.ContentHandler)
	mux.HandleFunc("/file/", a.loginProtect(a.FileHandler))
	mux.HandleFunc("/chat/", a.loginProtect(a.Chat))
	mux.HandleFunc("/chatep", a.loginProtect(a.ChatEndpoint))
	mux.HandleFunc("/github-api/", a.ProxyHandler)
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

func (a *App) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	apiURL := "https://api.github.com/" + r.URL.Path[len("/github-api/"):] + "?" + r.URL.RawQuery

	req, err := http.NewRequest(r.Method, apiURL, r.Body)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+a.GitHubAccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to forward request: %v", err)
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("GitHub API Error: %d %s", resp.StatusCode, resp.Status)
		http.Error(w, fmt.Sprintf("GitHub API Error: %d - %s", resp.StatusCode, resp.Status), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
