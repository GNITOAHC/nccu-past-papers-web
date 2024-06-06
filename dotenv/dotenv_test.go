package dotenv

import (
	"os"
	"testing"
)

func TestDotenv(t *testing.T) {
	err := Load("./env_test")
	t.Log("DOTENV_TEST_VAR_FIRST:", os.Getenv("DOTENV_TEST_VAR_FIRST"))
	t.Log("DOTENV_TEST:", os.Getenv("DOTENV_TEST"))
	t.Log("DOTENV_TEST_WITH_QUOTE:", os.Getenv("DOTENV_TEST_WITH_QUOTE"))
	t.Log("DOTENV_TEST_WITH_SPACE:", os.Getenv("DOTENV_TEST_WITH_SPACE"))
	t.Log("DOTENV_TEST_WITH_VAR:", os.Getenv("DOTENV_TEST_WITH_VAR"))
	if err != nil {
        t.Error("Error loading .env file:", err)
	} else {
		t.Log("Success loading .env file")
	}
}

func TestLogError(t *testing.T) {
	err := Load("./env_test_error")
	t.Log("Error:", err)
}
