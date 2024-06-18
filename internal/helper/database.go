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

func (h *Helper) Request(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	res, err := h.client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res, nil
}

func (h *Helper) getOneUser(mail string) []string {
	searchUrl := h.gasAPI + "?action=search&sheetName=past-papers-web-db&searchColumn=email&searchValue=" + mail
	res, err := h.Request("GET", searchUrl, nil)
	if err != nil {
		log.Print(err)
		return []string{}
	}
	body, err := io.ReadAll(res.Body)
	var data [][]string
	json.Unmarshal([]byte(body), &data)
	if len(data) == 0 {
		return []string{}
	}
	return data[0]
}

func (h *Helper) CheckUser(mail string) bool {
	data := h.getOneUser(mail)
	if len(data) == 0 {
		return false
	}
	if data[0] == mail {
		return true
	}
	return false
}

func (h *Helper) GetWaitingList() [][]string {
	res, err := h.Request("GET", h.gasAPI+"?action=readall&sheetName=waiting-list", nil)
	if err != nil {
		log.Print(err)
		return [][]string{}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return [][]string{}
	}

	var data [][]string
	json.Unmarshal([]byte(body), &data)
	return data
}

func (h *Helper) IsAdmin(mail string) bool {
	data := h.getOneUser(mail)
	if len(data) == 0 {
		return false
	}
	if data[0] == mail && data[5] == "true" {
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

	res, err := h.Request("POST", postUrl, strings.NewReader(reqBody))
	if err != nil {
		log.Print(err)
		return false
	}

	if res.Status != "200 OK" {
		return false
	}
	return true
}
