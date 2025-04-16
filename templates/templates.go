package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed all:*.html all:components/*.html all:footer/*.html
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

func _NewTemplates() {
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

// NewTemplates loads all templates from the templates directory and this must succeed
func NewTemplates() {
	root := template.New("")

	// Helper function to recursively process directories
	var processDir func(dir string) error
	processDir = func(dir string) error {
		entries, err := resources.ReadDir(dir)
		if err != nil {
			print("failed to read directory: ", dir, " ", err, "\n")
			panic(err)
		}

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				// Recursively process subdirectory
				if err := processDir(path); err != nil {
					panic(err)
				}
			} else if strings.HasSuffix(entry.Name(), ".html") {
				// Read and parse the template file
				content, err := resources.ReadFile(path)
				if err != nil {
					print("failed to read template file ", path, " ", err, "\n")
					panic(err)
				}

				t := root.New(path).Funcs(funcMap)
				_, err = t.Parse(string(content))
				if err != nil {
					print("failed to parse template ", path, " ", err, "\n")
					panic(err)
				}

				// print(path, " loaded\n")
			}
		}

		return nil
	}

	// Start processing from the root directory
	if err := processDir("."); err != nil {
		print("failed to load templates: ", err, "\n")
		panic(err)
	}

	tmpl = root
}

func Render(w http.ResponseWriter, name string, data interface{}) {
	var body bytes.Buffer

	// log.Println("Rendering template", name)

	// name = strings.Replace(name, "/", "_", -1)
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
