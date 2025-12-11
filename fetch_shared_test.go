package fetch

import (
	"os"
	"testing"

	"github.com/tinywasm/fetch"
)

func SendRequest_GetShared(t *testing.T, client fetch.Client) {
	done := make(chan bool)
	var responseBody []byte
	var responseErr error

	client.SendJSON("GET", "/get", nil, func(result []byte, err error) {
		if err != nil {
			responseErr = err
		} else {
			responseBody = result
		}
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, got %v", responseErr)
	}
	if string(responseBody) != "get success" {
		t.Errorf("Expected body 'get success', got '%s'", string(responseBody))
	}
}

func SendRequest_PostJSONShared(t *testing.T, client fetch.Client) {
	done := make(chan bool)
	requestData := map[string]string{"message": "hello"}
	var responseBody []byte
	var responseErr error

	client.SendJSON("POST", "/post_json", requestData, func(result []byte, err error) {
		if err != nil {
			responseErr = err
		} else {
			responseBody = result
		}
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, got %v", responseErr)
	}
	expected := `{"message":"hello"}`
	if string(responseBody) != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, string(responseBody))
	}
}

func SendRequest_TimeoutSuccessShared(t *testing.T, client fetch.Client) {
	done := make(chan bool)
	var responseErr error

	client.SendJSON("GET", "/timeout", nil, func(result []byte, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, but request timed out: %v", responseErr)
	}
}

func SendRequest_TimeoutFailureShared(t *testing.T, client fetch.Client) {
	done := make(chan bool)
	var responseErr error

	client.SendJSON("GET", "/timeout", nil, func(result []byte, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr == nil {
		t.Fatal("Expected request to time out, but it succeeded.")
	}
}

func SendRequest_ServerErrorShared(t *testing.T, client fetch.Client) {
	done := make(chan bool)
	var responseErr error

	client.SendJSON("GET", "/error", nil, func(result []byte, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr == nil {
		t.Fatal("Expected an error for 500 status code, but got nil.")
	}
}

func SendRequest_PostFileShared(t *testing.T, client fetch.Client) {
	// Create a temporary file with content.
	tmpfile, err := os.CreateTemp("", "test_upload_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "this is the content of the test file"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	done := make(chan bool)
	var responseBody []byte
	var responseErr error

	// Read file content and send as binary data.
	fileContent := []byte(content)
	client.SendBinary("POST", "/upload", fileContent, func(result []byte, err error) {
		if err != nil {
			responseErr = err
		} else {
			responseBody = result
		}
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error during file upload, got %v", responseErr)
	}
	if string(responseBody) != content {
		t.Errorf("Expected echoed file content '%s', got '%s'", content, string(responseBody))
	}
}
