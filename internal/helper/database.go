package helper

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

/*
 * Database schema: (Google Sheets; Columns should be placed in order in the first row of the sheet)
 * (1) sheetName: past-papers-web-db
 *     columns: email, name, studentId, contribute, time, admin
 *     - All columns are string
 *     - contribute/admin: "true" or "false"
 * (2) sheetName: waiting-list
 *     columns: email, name, studentId
 *     - All columns are string
 */

func (h *Helper) CheckUser(mail string) bool {
	searchUrl := h.gasAPI + "?action=search&sheetName=past-papers-web-db&searchColumn=email&searchValue=" + mail
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data [][]string
	json.Unmarshal([]byte(body), &data)

	if len(data) == 0 {
		return false
	}
	if data[0][0] == mail {
		return true
	}
	return false
}

func (h *Helper) GetWaitingList() [][]string {
	searchUrl := h.gasAPI + "?action=readall&sheetName=waiting-list"
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	var data [][]string
	json.Unmarshal([]byte(body), &data)
	return data
}

func (h *Helper) IsAdmin(mail string) bool {
	searchUrl := h.gasAPI + "?action=search&sheetName=past-papers-web-db&searchColumn=email&searchValue=" + mail
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchUrl, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data [][]string
	json.Unmarshal([]byte(body), &data)

	if len(data) == 0 {
		return false
	}
	if data[0][0] == mail && data[0][5] == "true" {
		return true
	}
	return false
}

func (h *Helper) RegisterUser(mail string, name string, studentId string) bool {
	postUrl := h.gasAPI
	userInfo := "[\"" + mail + "\", \"" + name + "\", \"" + studentId + "\"]"
	reqBody := `{
        "sheetName": "waiting-list",
        "action": "add",
        "record": ` + userInfo + `}`

	client := &http.Client{} // Create HTTP client
	req, err := http.NewRequest("POST", postUrl, strings.NewReader(reqBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	if res.Status != "200 OK" {
		return false
	}
	return true
}
