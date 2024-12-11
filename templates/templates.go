package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:*.html all:components/*.html
var resources embed.FS

var tmpl *template.Template = nil

// Define your template functions here
var funcMap = template.FuncMap{
	"dict": func(values ...interface{}) map[string]interface{} {
		if len(values)%2 != 0 {
			panic("dict function requires an even number of arguments")
		}
		result := make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				panic("dict function requires string keys")
			}
			result[key] = values[i+1]
		}
		return result
	},
	"list": func(values ...interface{}) []interface{} {
		return values
	},
	"split": func(s, sep string) []string {
		return strings.Split(s, sep)
	},
	"append": func(s []string, e string) []string {
		return append(s, e)
	},
	"slice": func() []string {
		return []string{}
	},
	// Add more functions as needed
}

func NewTemplates() {
	tmpl = template.New("").Funcs(funcMap)

	var paths []string
	fs.WalkDir(resources, ".", func(path string, d fs.DirEntry, err error) error {
		if strings.Contains(d.Name(), ".html") {
			paths = append(paths, path)
		}
		return nil
	})

	// log.Println(paths)

    // https://stackoverflow.com/questions/38686583/golang-parse-all-templates-in-directory-and-subdirectories

	tmpl = template.Must(tmpl.ParseFS(resources, paths...))
}

func Render(w http.ResponseWriter, name string, data interface{}) {
	var body bytes.Buffer

	// log.Println("Rendering template", name)

	name = strings.Replace(name, "/", "_", -1)
	err := tmpl.ExecuteTemplate(&body, name, data)
	if err != nil {
		err = fmt.Errorf("error executing template: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buffer bytes.Buffer

	err = tmpl.ExecuteTemplate(&buffer, "base.html", map[string]interface{}{
		"Body": template.HTML(body.String()),
	})
	if err != nil {
		err = fmt.Errorf("error executing template: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	buffer.WriteTo(w)
}
