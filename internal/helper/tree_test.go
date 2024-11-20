package helper

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestTreeParse(t *testing.T) {
	jsonFile, err := os.Open("./tree_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	body, err := io.ReadAll(jsonFile)

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	// mar, err := json.Marshal(data["tree"])
	// if err != nil {
	//     t.Log(err)
	// }
	// e := make([]githubEntry, 0)
	// json.Unmarshal(mar, &e)

	// for _, e := range(e) {
	//     t.Log(e.Path)
	// }

	root := ParseTree(data)
	printTree(root, 0)
	t.Log(root.Name)
}

func TestGetChildren(t *testing.T) {
	jsonFile, err := os.Open("./tree_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	body, err := io.ReadAll(jsonFile)
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	root := ParseTree(data)
	node, err := GetChildren(root, "ComputerScience/")
	if err != nil {
		t.Fatal(err)
	}
	printTree(node, 0)
}

func TestBFSSearch(t *testing.T) {
	jsonFile, err := os.Open("./tree_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	body, err := io.ReadAll(jsonFile)
	if err != nil {
		t.Fatal(err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatal(err)
	}

	root := ParseTree(data)

	// 測試搜尋包含 "alg" 的檔案和資料夾
	substring := "hw1"
	results := root.BFSSearch(substring)

	if len(results) == 0 {
		t.Log("未找到包含子字串:", substring)
	} else {
		for _, node := range results {
			t.Logf("找到: %s at path: %s\n", node.Name, node.Path)
		}
	}
}

func printTree(t *TreeNode, depth int) {
	isDir := t.IsDir
	if isDir {
		log.Println(strings.Repeat("  ", depth), t.Name+"/")
	} else {
		log.Println(strings.Repeat("  ", depth), t.Name, t.Size)
	}
	for _, c := range t.Children {
		printTree(c, depth+1)
	}
}
