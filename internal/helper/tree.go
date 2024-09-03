package helper

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"past-papers-web/internal/config"
)

type TreeNode struct {
	Name     string
	Path     string
	Size     int
	Children map[string]*TreeNode
	IsDir    bool
}

type githubEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
	Size *int   `json:"size,omitempty"` // pointer to an integer
	Url  string `json:"url"`
}

// InitTree initialize the tree node and returns the data from the GitHub API.
func InitTree(config *config.Config) map[string]interface{} {
	client := &http.Client{} // Create HTTP request
	req, err := http.NewRequest("GET", config.RepoAPI+"git/trees/main?recursive=1", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.GitHubAccessToken)
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	body, err := io.ReadAll(res.Body)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	return data
}

// Refresh the tree node at given helper instance by fetching the latest data from the GitHub API
func RefreshTree(config *config.Config, h *Helper) {
	h.TreeNode = ParseTree(InitTree(config))
	return
}

// GetChildren gets the children from the given path in the current tree.
func (t *TreeNode) GetChildren(path string) (*TreeNode, error) {
	if path == "" {
		return t, nil
	}
	return GetChildren(t, path)
}

// GetChildren gets the children of the given tree node and path.
func GetChildren(root *TreeNode, path string) (*TreeNode, error) {
	if path == "" {
		return root, nil
	}
	path = strings.TrimPrefix(strings.TrimSuffix(path, "/"), "/") // Trim both prefix "/" and suffix "/"
	paths := strings.Split(path, "/")
	current := root
	for _, p := range paths {
		if _, exists := current.Children[p]; !exists {
			return nil, errors.New("Path not found")
		}
		current = current.Children[p]
	}
	return current, nil
}

// AddPath adds a path to the tree.
func (t *TreeNode) AddPath(path []string, index int, entry githubEntry) {
	if index >= len(path) {
		return
	}
	current := path[index]
	// If the current node does not exist
	if _, exists := t.Children[current]; !exists {
		size := 0
		if entry.Size != nil {
			size = *entry.Size
		}
		t.Children[current] = &TreeNode{
			Name:     current,
			Path:     t.Path + "/" + current,
			Size:     size,
			Children: make(map[string]*TreeNode),
			IsDir:    entry.Type == "tree",
		}
	}
	t.Children[current].AddPath(path, index+1, entry)
}

// ParseTree parses the tree structure from the GitHub API response.
func ParseTree(data map[string]interface{}) *TreeNode {
	remar, err := json.Marshal(data["tree"]) // Re-marshal the data
	if err != nil {
		log.Println(err)
	}
	entry := make([]githubEntry, 0) // Convert to a slice of githubEntry
	json.Unmarshal(remar, &entry)

	// Define root node
	root := &TreeNode{
		Name:     "root",
		Path:     "",
		Size:     0,
		Children: make(map[string]*TreeNode),
		IsDir:    true,
	}

	for _, e := range entry {
		root.AddPath(strings.Split(e.Path, "/"), 0, e)
	}

	return root
}
