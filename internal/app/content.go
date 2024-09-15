package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"

	"past-papers-web/internal/helper"
)

type contentItem struct {
	Link     string
	Name     string
	Download string
	IsTree   bool
}

func (a *App) ContentHandler(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/content/"):]

	switch r.Method {
	case http.MethodPost:
		a.uploadFile(w, r)
		return
	case http.MethodGet:
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

func (a *App) GetContent(w http.ResponseWriter, r *http.Request, urlpath string) {
	items := make([]contentItem, 0)

	treeNode, err := a.helper.TreeNode.GetChildren(urlpath)
	if err != nil {
		fmt.Println("Error getting children", err)
	}

	keys := make([]string, 0, len(treeNode.Children)) // Sort keys
	for k := range treeNode.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := treeNode.Children[k]
		link := ""
		if v.IsDir {
			link = "/content" + v.Path
		} else {
			link = "/chat" + v.Path
		}
		items = append(items, contentItem{
			Link:     link, // Absolute path (starts with /content)
			Name:     v.Name,
			Download: "/file" + v.Path,
			IsTree:   v.IsDir,
		})
	}

	// Prepend slash to url if not empty
	if urlpath != "" {
		urlpath = "/" + urlpath
	}
	a.tmplExecute(w, []string{"templates/content.html"}, map[string]interface{}{
		"Path":  "content" + urlpath,
		"Items": items,
	})
	return
}
