package app

import (
	"flag"
	"fmt"
	"html/template"
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
	mux.HandleFunc("/content/", a.loginProtect(a.ContentHandler))
	return mux
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	renderTmpl := func() {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
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

func (a *App) GetStructure(w http.ResponseWriter, r *http.Request) {
	res := a.helper.GetStructure()
	tmpl := template.Must(template.ParseFiles("templates/tree.html"))

	type TmplTree struct {
		Path string
		Name string
	}
	tmplTree := make([]TmplTree, 0)

	if tree, ok := res["tree"].([]interface{}); ok {
		for _, item := range tree {
			if treeItem, ok := item.(map[string]interface{}); ok {
				if path, ok := treeItem["path"].(string); ok && treeItem["type"].(string) == "tree" {
					// fmt.Println("Path:", path)
					tmplTree = append(tmplTree, TmplTree{Path: "content/" + path, Name: path})
				}
			}
		}
	}

	content := map[string]interface{}{
		"Tree":  tmplTree,
		"Title": "Tree",
	}
	err := tmpl.Execute(w, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}
