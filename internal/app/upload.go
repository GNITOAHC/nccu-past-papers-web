package app

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"past-papers-web/internal/helper"
)

func getTime() string {
	return time.Now().Format("01-02-15-04")
}

func (a *App) uploadSingleFile(path, filename string, buf *bytes.Buffer, id int) error {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(buf.Bytes())))
	base64.StdEncoding.Strict().Encode(dst, buf.Bytes())

	newBranchName := "upload-" + getTime() + "-" + strconv.Itoa(id)
	newBranchSha, err := a.helper.CreateBranch(newBranchName)
	if err != nil {
		return errors.New("Error creating branch: " + err.Error())
	}

	uploadData := helper.UploadData{
		Message: "Uploaded from web",
		Content: string(dst),
		Branch:  newBranchName,
		Sha:     newBranchSha,
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	err = a.helper.Upload(&uploadData, path+filename)
	if err != nil {
		return errors.New("Error uploading file: " + err.Error())
	}

	err = a.helper.CreatePR(newBranchName)
	if err != nil {
		return errors.New("Error creating PR " + err.Error())
	}

	return nil
}

func (a *App) uploadFiles(w http.ResponseWriter, r *http.Request) {
	// Function scope variables.
	var (
		p    *multipart.Part // for getting file info at end
		err  error
		path string
	)

	queue := []struct {
		filename string
		buf      *bytes.Buffer
	}{}

	mr, err := r.MultipartReader()
	if err != nil {
		print("Hit error while opening multipart reader: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for {
		p, err = mr.NextPart()
		if err == io.EOF {
			// err is io.EOF, files upload completes.
			print("Hit last part of multipart upload\n")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Files upload complete"))
			break
		}
		if err != nil {
			// A normal error occurred
			print("Hit error while fetching next part: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, p); err != nil {
			print("Error copying part:", err)
			return
		}

		if p.FormName() == "path" {
			path = buf.String()
			continue
		}

		if p.FormName() == "file" {
			// This is a file to upload.
			queue = append(queue, struct {
				filename string
				buf      *bytes.Buffer
			}{p.FileName(), &buf})
		}
	}

	print("Path to upload to: ", path, "\n")
	for idx, q := range queue {
		print("Uploading file: ", q.filename, "\n")
		err := a.uploadSingleFile(path, q.filename, q.buf, idx)
		if err != nil {
			print("Error uploading file: ", err, "\n")
			w.Write([]byte("Error uploading file: " + err.Error()))
			return
		}
	}
}
