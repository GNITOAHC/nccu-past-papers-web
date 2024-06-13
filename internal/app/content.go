package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

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
		http.Error(w, "Error getting file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		http.Error(w, "Error copying file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(buf.Bytes())))
	base64.StdEncoding.Strict().Encode(dst, buf.Bytes())

	newBranchName := "upload-from-" + name
	prefix := len("upload-from-" + name)
	var newBranchSha string
	for i := 0; i < 10; i++ {
		newBranchSha, err = a.helper.CreateBranch(newBranchName)
		if err == nil {
			break
		}
		newBranchName = newBranchName[:prefix] + "-" + fmt.Sprint(i)
	}
	if err != nil {
		http.Error(w, "Error creating branch: "+err.Error(), http.StatusInternalServerError)
	}

	uploadData := helper.UploadData{
		Message: "Upload by " + name,
		Content: string(dst),
		Branch:  newBranchName,
		Sha:     newBranchSha,
	}
	err = a.helper.Upload(&uploadData, r.URL.Path[len("/content/"):]+"/"+header.Filename)
	if err != nil {
		http.Error(w, "Error uploading file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.helper.CreatePR(newBranchName)
	if err != nil {
		w.Write([]byte("Error creating PR " + err.Error()))
		return
	}

	w.Write([]byte("File Uploaded"))
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
	items := make([]contentItem, 0)

	treeNode, err := a.helper.TreeNode.GetChildren("")
	if err != nil {
		fmt.Println("Error getting children", err)
	}
	for _, v := range treeNode.Children {
		items = append(items, contentItem{
			Link:   "/content" + v.Path, // Absolute path (starts with /content)
			Name:   v.Name,
			IsTree: v.IsDir,
		})
	}

	a.tmplExecute(w, map[string]interface{}{
		"Title": "Path >>> HOME",
		"Items": items,
	})
	return
}

func (a *App) GetContent(w http.ResponseWriter, r *http.Request, urlpath string) {
	items := make([]contentItem, 0)

	treeNode, err := a.helper.TreeNode.GetChildren(urlpath)
	if err != nil {
		fmt.Println("Error getting children", err)
	}
	for _, v := range treeNode.Children {
		items = append(items, contentItem{
			Link:   "/content" + v.Path, // Absolute path (starts with /content)
			Name:   v.Name,
			IsTree: v.IsDir,
		})
	}

	a.tmplExecute(w, map[string]interface{}{
		"Title": "Path >>> " + urlpath,
		"Items": items,
	})
	return
}
