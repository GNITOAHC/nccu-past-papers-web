package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	
	"past-papers-web/internal/config"
)

// Helper struct contains all the necessary data for the helper functions.
type Helper struct {
	githubAccessToken string
	repoAPI           string
	authorization     string
	gasAPI            string
	client            *http.Client
	TreeNode          *TreeNode // TreeNode: Root node of the tree structure.
}

// NewHelper creates a new helper instance.
func NewHelper(config *config.Config) *Helper {
	treeNode := ParseTree(InitTree(config))
	return &Helper{
		githubAccessToken: config.GitHubAccessToken,
		repoAPI:           config.RepoAPI,
		authorization:     "Bearer " + config.GitHubAccessToken,
		gasAPI:            config.GASAPI,
		client:            &http.Client{},
		TreeNode:          treeNode,
	}
}

type FileChange struct {
	Filename   string `json:"filename"`
	Number int    `json:"number"`
}

type PullRequest struct {
	Number int    `json:"number"`
}

func (a *Helper) GetPullRequest(repoOwner string, repoName string, token string) ([]PullRequest, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", repoOwner, repoName)
	req, err := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+ token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Print(errors.New("failed to fetch pull requests, status: " + resp.Status))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var prs []PullRequest
	err = json.Unmarshal(body, &prs)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return prs, nil
}

func (a *Helper) GetFileChange(repoOwner string, repoName string, PRnumber int, token string) ([]FileChange, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", repoOwner, repoName, PRnumber)
	req, err := http.NewRequest("GET", apiUrl, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+ token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Print(errors.New("failed to fetch file change, status: " + resp.Status))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var fls []FileChange
	err = json.Unmarshal(body, &fls)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return fls, nil
}

func (a *Helper) GetPRList(repoOwner string, repoName string, token string) []FileChange {
	prNumber, err := a.GetPullRequest(repoOwner, repoName, token)
	if err != nil {
		log.Fatalf("Error fetching pull requests: %v", err)
	}
	var result []FileChange
	for _, pr := range prNumber {
		files, err := a.GetFileChange(repoOwner, repoName, pr.Number, token)
		if err != nil {
			log.Fatalf("Error fetching file change for Pull Request #%d: %v", pr.Number, err)
			continue
		}

		for _, file := range files{
			list := FileChange{
				file.Filename,
				pr.Number,
			}
			result = append(result, list)
		}
	}
	return result
}

// Send a request given the method, URL, body and header. Returning the raw response and error.
func (h *Helper) request(method, url string, body io.Reader, header map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := h.client.Do(req)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return res, nil
}

// Get the SHA of the main branch. Returning the SHA of the main branch.
func (h *Helper) GetSHA() string {
	// https://docs.github.com/en/rest/git/refs?apiVersion=2022-11-28#get-a-reference
	// Repository permissions for "Contents"
	// Repository permissions for "Workflows"
	res, err := h.request("GET", h.repoAPI+"git/refs/heads/main", nil, map[string]string{"Authorization": h.authorization})
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	body, err := io.ReadAll(res.Body)
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	return data["object"].(map[string]interface{})["sha"].(string)
}

// CreateBranch creates a branch with the given name and returns the SHA of the created branch
func (h *Helper) CreateBranch(branch string) (string, error) {
	// https://docs.github.com/en/rest/git/refs?apiVersion=2022-11-28#create-a-reference
	// Repository permissions for "Contents"
	// Repository permissions for "Workflows"
	sha := h.GetSHA()
	jsonData := []byte(`{"ref": "refs/heads/` + branch + `", "sha": "` + sha + `"}`)

	res, err := h.request("POST", h.repoAPI+"git/refs", bytes.NewBuffer(jsonData), map[string]string{"Authorization": h.authorization})
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

// UploadData contains essential information for uploading a file to the repository.
type UploadData struct {
	Message string
	Content string // Base64 encoded string
	Branch  string
	Sha     string
}

// UploadData uploads the given data to the repository at the given path
func (h *Helper) Upload(uploadData *UploadData, path string) error {
	// https://docs.github.com/en/rest/repos/contents?apiVersion=2022-11-28#create-or-update-file-contents
	// Repository permissions for "Contents"
	header := map[string]string{
		"Authorization": h.authorization,
		"Accept":        "application/vnd.github+json",
	}
	jsonStr := []byte(`{
        "message": "` + uploadData.Message + `", ` +
		`"content": "` + uploadData.Content + `", ` +
		`"branch": "` + uploadData.Branch + `", ` +
		`"sha": "` + uploadData.Sha + `"}`)

	res, err := h.request("PUT", h.repoAPI+"contents/"+strings.TrimPrefix(path, "/"), bytes.NewBuffer(jsonStr), header)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	if res.StatusCode == 201 {
		return nil
	}
	return errors.New(data["message"].(string) + ", Status Code: " + strconv.Itoa(res.StatusCode))
}

// CreatePR creates a pull request with the given name
func (h *Helper) CreatePR(branch string) error {
	// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#create-a-pull-request
	// Repository permissions for "Pull requests"
	header := map[string]string{
		"Authorization": h.authorization,
		"Accept":        "application/vnd.github+json",
	}
	jsonStr := []byte(`{"head": "gnitoahc:` + branch + `", "base": "main", "title": "Create PR test"}`)

	res, err := h.request("POST", h.repoAPI+"pulls", bytes.NewBuffer(jsonStr), header)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	if res.StatusCode == 201 {
		return nil
	}
	return errors.New(data["message"].(string) + ", Status Code: " + strconv.Itoa(res.StatusCode))
}

// GetFile returns the raw content of the file
func (h *Helper) GetFile(path string) ([]byte, error) {
	// https://docs.github.com/en/rest/repos/contents?apiVersion=2022-11-28#get-repository-content
	// Repository permissions for "Contents"
	// Repository permissions for "Workflows"
	header := map[string]string{
		"Authorization": h.authorization,
		"Accept":        "application/vnd.github.raw+json",
	}
	res, err := h.request("GET", h.repoAPI+"contents/"+path, nil, header)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	return body, nil
}

func (h *Helper) FileReader(path string) (io.Reader, error) {
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

    return res.Body, nil
}
