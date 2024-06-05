package helper

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
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

func (t *TreeNode) GetChildren(path string) (*TreeNode, error) {
	if path == "" {
		return t, nil
	}
	return GetChildren(t, path)
}

func GetChildren(root *TreeNode, path string) (*TreeNode, error) {
	if path == "" {
		return root, nil
	}
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
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
