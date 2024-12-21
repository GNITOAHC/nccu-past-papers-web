package app

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (a *App) FileHandler(w http.ResponseWriter, r *http.Request) {
	urlpath := r.URL.Path[len("/file/"):]
	fileContent := map[string]string{
		".pdf":  "application/pdf",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".txt":  "text/plain; charset=utf-8",
		".md":   "text/plain; charset=utf-8",
	}

	contentType := ""
	for k, v := range fileContent { // If GET request for file
		if strings.HasSuffix(urlpath, k) {
			contentType = v
			break
		}
	}
	if contentType == "" {
		// http.Error(w, "Content type not specified or supported", http.StatusNotFound)
		// return
		contentType = "application/octet-stream"
	}

	var file []byte
	if data, has := a.filecache.Get(urlpath); has { // Has file in cache
		file = data
	} else {
		data, err := a.helper.GetFile(urlpath)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
		}
		a.filecache.Set(urlpath, data, time.Duration(time.Hour*360)) // Cache for 15 days
		file = data
	}

	b := bytes.NewBuffer(file)

	w.Header().Set("Content-type", contentType)
	if _, err := b.WriteTo(w); err != nil {
		fmt.Fprintf(w, "%s", err)
	}

	w.Write([]byte("File Generated"))
	return
}
