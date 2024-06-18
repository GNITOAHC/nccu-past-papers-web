package app

import (
	"net/http"
	"text/template"
)

func (a *App) tmplExecute(w http.ResponseWriter, tmpls []string, data map[string]interface{}) {
    tmpls = append([]string{"templates/base.html"}, tmpls...)
	tmpl := template.Must(template.ParseFiles(tmpls...))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
