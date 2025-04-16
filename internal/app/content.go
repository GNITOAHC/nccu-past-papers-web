package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"past-papers-web/internal/helper"
	"past-papers-web/templates"
)

type contentItem struct {
	Link     string
	Name     string
	Download string
	IsTree   bool
}

func (a *App) DownloadZip(w http.ResponseWriter, r *http.Request) {
	userMail, err := r.Cookie("email")
	log.Println("Download zip request from user: ", userMail.Value)

	url := "https://api.github.com/repos/GNITOAHC/nccu-past-papers/zipball/main"

	// print("DownloadZip function called")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+a.config.GitHubAccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Error downloading zip: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		http.Error(w, "Error writing zip to response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=past-papers-archieve.zip")
	w.Header().Set("Content-Type", "application/zip")
	w.Write(buf.Bytes())
}

func (a *App) ContentHandler(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/content/"):]

	switch r.Method {
	case http.MethodPost:
		a.uploadFile(w, r)
		return
	case http.MethodGet:
		if strings.HasSuffix(urlpath, "/download-zip") {
			a.DownloadZip(w, r)
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

func (a *App) GetContent(w http.ResponseWriter, r *http.Request, urlpath string) {
	// Define files to be ignored
	ignores := []string{"README.md", ".gitignore"}

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

	hasreadme := false

outer:
	for _, k := range keys {
		v := treeNode.Children[k]
		link := ""
		if v.IsDir {
			link = "/content" + v.Path
		} else {
			link = "/chat" + v.Path
		}
		if v.Name == "README.md" {
			hasreadme = true
		}

		// Ignore files
		for _, ignore := range ignores {
			if v.Name == ignore {
				continue outer
			}
		}
		items = append(items, contentItem{
			Link:     link, // Absolute path (starts with /content)
			Name:     v.Name,
			Download: "/file" + v.Path,
			IsTree:   v.IsDir,
		})
	}

	var readme string
	if hasreadme {
		readmeb, err := a.helper.GetFile(urlpath + "/README.md")
		if err != nil {
			fmt.Println("Error getting readme", err)
		}
		readme = string(readmeb)
	}

	// Prepend slash to url if not empty
	if urlpath != "" {
		urlpath = "/" + urlpath
	}
	// a.tmplExecute(w, []string{"templates/content.html"}, map[string]interface{}{
	// 	"Path":      "content" + urlpath,
	// 	"Items":     items,
	// 	"HasReadme": hasreadme,
	// 	"Readme":    readme,
	// })
	// log.Println(items)
	templates.Render(w, "content.html", map[string]interface{}{
		"Path":      "content" + urlpath,
		"Items":     items,
		"HasReadme": hasreadme,
		"Readme":    readme,
	})
	// return
}
