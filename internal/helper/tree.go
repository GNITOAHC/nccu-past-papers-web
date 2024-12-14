package helper

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"past-papers-web/internal/config"
)

type TreeNode struct {
	Name     string               `json:"name"`
	Path     string               `json:"path"`
	Size     int                  `json:"size"`
	Children map[string]*TreeNode `json:"children,omitempty"`
	IsDir    bool                 `json:"isDir"`
}

type githubEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
	Size *int   `json:"size,omitempty"` // pointer to an integer
	Url  string `json:"url"`
}

type ProxyHelper struct {
	Client *http.Client
	Token  string
}

// BFSSearch performs a breadth-first search for files or directories that contain the specified substring, case insensitive.
func (t *TreeNode) BFSSearch(substring string) []*TreeNode {
	var results []*TreeNode
	queue := []*TreeNode{t} // Use a queue for BFS

	// Convert the search substring to lowercase
	substringLower := strings.ToLower(substring)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:] // Remove the processed element

		// Compare the current node's name and path, converted to lowercase
		if strings.Contains(strings.ToLower(current.Name), substringLower) ||
			strings.Contains(strings.ToLower(current.Path), substringLower) {
			results = append(results, current)
		}

		// Add child nodes to the queue
		for _, child := range current.Children {
			queue = append(queue, child)
		}
	}

	return results
}

// Handle search requests
func SearchHandler(w http.ResponseWriter, r *http.Request, root *TreeNode) {
	// Retrieve the search string from the URL query parameters
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	// Perform BFS
	results := root.BFSSearch(query)

	// Return search results
	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "No matching results found"})
		return
	}

	// Return JSON
	json.NewEncoder(w).Encode(results)
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

func NewProxyHelper() *ProxyHelper {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN is not set in environment variables")
	}
	return &ProxyHelper{
		Client: &http.Client{},
		Token:  token,
	}
}

// HandleProxyRequest proxies GitHub API requests
func (p *ProxyHelper) HandleProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Construct the GitHub API request URL
	apiURL := "https://api.github.com" + r.URL.Path[len("/github-api"):] + "?" + r.URL.RawQuery

	// Create a new request
	req, err := http.NewRequest(r.Method, apiURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send the request
	resp, err := p.Client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Set response headers and return data
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
