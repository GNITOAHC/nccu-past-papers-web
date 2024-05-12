package app

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	"past-papers-web/internal/config"
	"past-papers-web/internal/helper"
)

var (
	port = flag.Int("port", 3000, "The server port")
)

type App struct {
	helper *helper.Helper
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
	return &App{
		helper: helper.NewHelper(config),
	}
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", a.middlewares(http.HandlerFunc(a.Login)))
	mux.Handle("/tree", a.middlewares(http.HandlerFunc(a.GetStructure)))
	mux.Handle("/content/", a.middlewares(http.HandlerFunc(a.HandleContent)))
	return mux
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	renderTmpl := func() {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if r.Method == "POST" {
		email := r.FormValue("email")
		// Set a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "email",
			Value:    email,
			MaxAge:   3600,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		if a.helper.CheckUser(email) { // Has user in DB
			http.Redirect(w, r, "/tree", http.StatusSeeOther)
			return
		}

		renderTmpl()
		return
	}
	renderTmpl()
	return
}

func (a *App) middlewares(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}
		// Check if the user is logged in
		cookie, err := r.Cookie("email")
		if err != nil { // Error occuring while getting cookie
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if a.helper.CheckUser(cookie.Value) { // Has user in DB
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
