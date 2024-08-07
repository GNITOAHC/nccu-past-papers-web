package helper

import (
	"encoding/json"
	"errors"
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

// Database names
const (
	RegisterDB    = "past-papers-web-db"
	WaitingListDB = "waiting-list"
)

// request sends a request to the server according to the given parameters.
//
// @param method: HTTP method, _e.g._ "GET", "POST"
// @param url: URL to send the request
// @param body: Request body
//
// @return *http.Response: Response from the server
// @return error: Returns an error if the request fails.
func (h *Helper) request(method string, url string, body io.Reader) (*http.Response, error) {
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

// getOneUser returns the user data from the database.
//
// @param mail: mail to check
//
// @return []string: user data _e.g._ ["mail@mail.com", "GNITOAHC", "123456"]
//
// If the user does not exists, it returns an empty array.
func (h *Helper) getOneUser(mail string) []string {
	searchUrl := h.gasAPI + "?action=search&sheetName=past-papers-web-db&searchColumn=email&searchValue=" + mail
	res, err := h.request("GET", searchUrl, nil)
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

// CheckUser reports whether the user exists in the database.
//
// @param mail: mail to check
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

// ApproveRegistration approves the registration of the user.
//
// @param mail, name, studentId: user's data
//
// @return error: Returns an error if the registration fails.
func (h *Helper) ApproveRegistration(mail, name, studentId string) error {
	deleteBody := `{
        "sheetName": "waiting-list",
        "action": "delete",
        "columnName": "email",
        "rowValue": "` + mail + `"}`
	res, err := h.request("POST", h.gasAPI, strings.NewReader(deleteBody))
	if err != nil {
		log.Print(err)
		return err
	}
	_, err = io.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return err
	}

	registered := h.CheckUser(mail)
	if !registered {
		userInfo := "[\"" + mail + "\", \"" + name + "\", \"'" + studentId + "\"]"
		reqBody := `{
            "sheetName": "past-papers-web-db",
            "action": "add",
            "record": ` + userInfo + `}`

		res, err := h.request("POST", h.gasAPI, strings.NewReader(reqBody))
		if err != nil {
			log.Print(err)
			return err
		}
		if res.Status != "200 OK" {
			return errors.New("Failed to add user")
		}
	}
	return nil
}

// GetWaitingList returns the waiting list from the database.
//
// @return [][]string: Waiting list data _e.g._ [["mail1", "name1", "123456"], ["mail2", "name2", "654321"]]
func (h *Helper) GetWaitingList() [][]string {
	res, err := h.request("GET", h.gasAPI+"?action=readall&sheetName=waiting-list", nil)
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

// IsAdmin reports whether the user is an admin.
//
// @param mail: mail to check
//
// @return bool: Returns true if the user is an admin.
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

// RegisterUser registers the user to the waiting list.
//
// @param mail, name, studentId: user's data
//
// @return bool: Returns true if the registration is successful.
func (h *Helper) RegisterUser(mail string, name string, studentId string) bool {
	userInfo := "[\"" + mail + "\", \"" + name + "\", \"'" + studentId + "\"]"
	reqBody := `{
        "sheetName": "waiting-list",
        "action": "add",
        "record": ` + userInfo + `}`

	res, err := h.request("POST", h.gasAPI, strings.NewReader(reqBody))
	if err != nil {
		log.Print(err)
		return false
	}

	if res.Status != "200 OK" {
		return false
	}
	return true
}
