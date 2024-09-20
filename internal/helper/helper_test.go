package helper

import (
	b64 "encoding/base64"
	"os"
	"testing"

	"past-papers-web/internal/config"
)

func TestUploadCombos(t *testing.T) {
	config := config.NewConfig("../../.env")
	h := NewHelper(config)

	newBranchName := "upload-test-0"
	newBranchSha, err := h.CreateBranch(newBranchName)
	if err != nil {
		t.Log(err)
	}

	content, err := os.ReadFile("test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	dst := make([]byte, b64.StdEncoding.EncodedLen(len(content)))
	b64.StdEncoding.Strict().Encode(dst, content)

	// data := "abc123!?$*&()'-=@~"

	t.Log(newBranchSha)

	uploadData := UploadData{
		Message: "Test upload",
		// content: b64.RawStdEncoding.Strict().Encode(content),
		// content: b64.RawStdEncoding.EncodeToString([]byte(data)),
		Content: string(dst),
		Branch:  newBranchName,
		Sha:     newBranchSha,
	}

	// t.Log(newBranchSha)

	err = h.Upload(&uploadData, "test.pdf")
	if err != nil {
		t.Log(err)
	}

	err = h.CreatePR(newBranchName)
	if err != nil {
		t.Log(err)
	}

	return
}
