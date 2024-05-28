package app

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"encoding/base64"

	"past-papers-web/internal/helper"
)

type contentItem struct {
	Link   string
	Name   string
	IsTree bool
}

func (a *App) ContentHandler(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/content/"):]
	fileContent := map[string]string{
		".pdf":  "application/pdf",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
	}

	switch r.Method {
	case http.MethodPost:
		a.uploadFile(w, r)
		return
	case http.MethodGet:
		for k, v := range fileContent { // If GET request for file
			if strings.HasSuffix(urlpath, k) {
				a.handleFile(w, v, urlpath)
				return
			}
		}
		if urlpath == "" {
			a.GetHomeContent(w, r) // Handle home page
			return
		}
		a.GetContent(w, r, urlpath) // Else handle content page
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (a *App) uploadFile(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error getting file", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(buf.Bytes())))
	base64.StdEncoding.Strict().Encode(dst, buf.Bytes())

	newBranchName := "upload-from-" + name
	newBranchSha, err := a.helper.CreateBranch(newBranchName)
	if err != nil {
		fmt.Println("Error creating branch", err)
	}

	uploadData := helper.UploadData{
		Message: "Upload from " + name,
		Content: string(dst),
		Branch:  newBranchName,
		Sha:     newBranchSha,
	}
	err = a.helper.Upload(&uploadData, r.URL.Path[len("/content/"):]+"/"+header.Filename)
	if err != nil {
		fmt.Println("Error uploading file", err)
	}

	err = a.helper.CreatePR(newBranchName)
	if err != nil {
		fmt.Println("Error creating PR", err)
	}

	tmpl := template.Must(template.ParseFiles("templates/success.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Redirect": r.URL.Path[len("/content/"):],
	})
	return
}

func (a *App) handleFile(w http.ResponseWriter, contentType string, urlpath string) {
	var pdfData []byte
	if data, has := a.filecache.Get(urlpath); has { // Has file in cache
		pdfData = data
	} else {
		data, err := a.helper.GetFile(urlpath)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
		}
		a.filecache.Set(urlpath, data, time.Duration(time.Hour*360)) // Cache for 15 days
		pdfData = data
	}

	b := bytes.NewBuffer(pdfData)

	w.Header().Set("Content-type", contentType)
	if _, err := b.WriteTo(w); err != nil {
		fmt.Fprintf(w, "%s", err)
	}

	w.Write([]byte("File Generated"))
	return
}

func (a *App) tmplExecute(w http.ResponseWriter, data map[string]interface{}) {
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/content.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) GetHomeContent(w http.ResponseWriter, r *http.Request) {
	res := a.helper.GetStructure()
	items := make([]contentItem, 0)

	for _, v := range res["tree"].([]interface{}) {
		if treeItem, ok := v.(map[string]interface{}); ok {
			if path, ok := treeItem["path"].(string); ok {
				if len(strings.Split(path, "/")) == 1 {
					items = append(items, contentItem{
						Link:   path,
						Name:   path,
						IsTree: treeItem["type"].(string) == "tree",
					})
				}
			}
		}
	}

	a.tmplExecute(w, map[string]interface{}{
		"Title": "Path >>> HOME",
		"Items": items,
	})
	return
}

func (a *App) GetContent(w http.ResponseWriter, r *http.Request, urlpath string) {
	res := a.helper.GetStructure()
	items := make([]contentItem, 0)

	for _, v := range res["tree"].([]interface{}) {
		if treeItem, ok := v.(map[string]interface{}); ok {
			if path, ok := treeItem["path"].(string); ok {
				if strings.HasPrefix(path, urlpath) && len(strings.Split(path, "/")) == len(strings.Split(urlpath, "/"))+1 {
					lnk := strings.Split(urlpath, "/")[len(strings.Split(urlpath, "/"))-1] + "/" + strings.Split(path, "/")[len(strings.Split(urlpath, "/"))]
					items = append(items, contentItem{
						Link:   lnk,
						Name:   lnk,
						IsTree: treeItem["type"].(string) == "tree",
					})
				}
			}
		}
	}

	a.tmplExecute(w, map[string]interface{}{
		"Title": "Path >>> " + urlpath,
		"Items": items,
	})
	return
}
