package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"past-papers-web/internal/cache"
	"past-papers-web/internal/config"
)

type Helper struct {
	githubAccessToken string
	repoAPI           string
	authorization     string
	gasAPI            string
	structureCache    *cache.Cache[string, map[string]interface{}]
}

func NewHelper(config *config.Config) *Helper {
	structureCache := cache.New[string, map[string]interface{}]()
	return &Helper{
		githubAccessToken: config.GitHubAccessToken,
		repoAPI:           config.RepoAPI,
		authorization:     "Bearer " + config.GitHubAccessToken,
		gasAPI:            config.GASAPI,
		structureCache:    structureCache,
	}
}

// Repo structucture
func (h *Helper) GetStructure() map[string]interface{} {
	if val, ok := h.structureCache.Get("structure"); ok {
		return val
	}
	// Create HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.repoAPI+"git/trees/main?recursive=1", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", h.authorization)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	// fmt.Println(string(body))

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	// fmt.Println(data)
	h.structureCache.Set("structure", data, 7*24*time.Hour) // Cache for 7 days

	return data
}

func (h *Helper) GetSHA(apiPrefix string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiPrefix+"/heads/main", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", h.authorization)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	return data["object"].(map[string]interface{})["sha"].(string)
}

func (h *Helper) CreateBranch(branch string) (string, error) {
	branchAPI := h.repoAPI + "git/refs"
	sha := h.GetSHA(branchAPI)
	client := &http.Client{}
	jsonData := []byte(`{"ref": "refs/heads/` + branch + `", "sha": "` + sha + `"}`)

	req, err := http.NewRequest("POST", branchAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", h.authorization)

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	if res.StatusCode == 201 {
		return data["object"].(map[string]interface{})["sha"].(string), nil
	}
	return "", errors.New(data["message"].(string))
}

type UploadData struct {
	Message string
	Content string
	Branch  string
	Sha     string
}

func (h *Helper) Upload(uploadData *UploadData, path string) error {
	client := &http.Client{}
	jsonStr := []byte(`{
        "message": "` + uploadData.Message + `", ` +
		`"content": "` + uploadData.Content + `", ` +
		`"branch": "` + uploadData.Branch + `", ` +
		`"sha": "` + uploadData.Sha + `"}`)

	req, err := http.NewRequest("PUT", h.repoAPI+"contents/"+path, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", h.authorization)
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	if res.StatusCode == 201 {
		return nil
	}
	return errors.New(data["message"].(string) + ", Status Code: " + strconv.Itoa(res.StatusCode))
}

func (h *Helper) CreatePR(branch string) error {
	client := &http.Client{}
	jsonStr := []byte(`{"head": "gnitoahc:` + branch + `", "base": "main", "title": "Create PR test"}`)

	req, err := http.NewRequest("POST", h.repoAPI+"pulls", bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", h.authorization)
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	if res.StatusCode == 201 {
		return nil
	}
	return errors.New(data["message"].(string) + ", Status Code: " + strconv.Itoa(res.StatusCode))
}

func (h *Helper) GetFile(path string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.repoAPI+"contents/"+path, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", h.authorization)
	req.Header.Set("Accept", "application/vnd.github.raw+json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	return body, nil
}

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
