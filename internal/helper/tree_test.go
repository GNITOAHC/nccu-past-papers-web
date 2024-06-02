package helper

import (
	"encoding/json"
	"io"
	"log"
	"os"
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
	printTree(root)
	t.Log(root.Name)
}

func printTree(t *TreeNode) {
	log.Println(t.Name, t.Path)
	for _, c := range t.Children {
		printTree(c)
	}
}
