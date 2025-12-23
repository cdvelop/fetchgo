package fetch_test

import (
	"testing"

	"github.com/tinywasm/fetch"
)

func SendRequest_GetShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var responseBody string
	var responseErr error

	fetch.Get(baseURL + "/get").Send(func(resp *fetch.Response, err error) {
		if err != nil {
			responseErr = err
		} else {
			responseBody = resp.Text()
		}
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, got %v", responseErr)
	}
	if responseBody != "get success" {
		t.Errorf("Expected body 'get success', got '%s'", responseBody)
	}
}

func SendRequest_PostJSONShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	requestData := `{"message":"hello"}`
	var responseBody string
	var responseErr error

	fetch.Post(baseURL+"/post_json").
		Header("Content-Type", "application/json").
		Body([]byte(requestData)).
		Send(func(resp *fetch.Response, err error) {
			if err != nil {
				responseErr = err
			} else {
				responseBody = resp.Text()
			}
			done <- true
		})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, got %v", responseErr)
	}
	// The server should reflect the JSON we sent.
	// Since we are sending raw bytes, we expect exact match if server behaves simply,
	// but JSON serialization might slightly differ in spacing if we used a library.
	// Here we used a string literal, and server likely decodes/encodes.
	// Let's assume the server returns `{"message":"hello"}`.
	expected := `{"message":"hello"}`
	if responseBody != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, responseBody)
	}
}

func SendRequest_TimeoutSuccessShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var responseErr error

	fetch.Get(baseURL+"/timeout").
		Timeout(2000). // 2 seconds should be enough for the /timeout endpoint (usually 100ms or so in tests)
		Send(func(resp *fetch.Response, err error) {
			responseErr = err
			done <- true
		})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, but request timed out: %v", responseErr)
	}
}

func SendRequest_TimeoutFailureShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var responseErr error

	fetch.Get(baseURL+"/timeout").
		Timeout(10). // 10ms should be too short
		Send(func(resp *fetch.Response, err error) {
			responseErr = err
			done <- true
		})

	<-done

	if responseErr == nil {
		t.Fatal("Expected request to time out, but it succeeded.")
	}
}

func SendRequest_ServerErrorShared(t *testing.T, baseURL string) {
	done := make(chan bool)
	var status int
	var responseErr error

	fetch.Get(baseURL + "/error").Send(func(resp *fetch.Response, err error) {
		if err != nil {
			responseErr = err
		} else {
			status = resp.Status
		}
		done <- true
	})

	<-done

	// In the new API, 500 is not an error in the callback sense (network error),
	// it's a valid response with status 500.
	if responseErr != nil {
		t.Fatalf("Expected no network error, got %v", responseErr)
	}
	if status != 500 {
		t.Errorf("Expected status 500, got %d", status)
	}
}

func SendRequest_PostFileShared(t *testing.T, baseURL string) {
	// Create a temporary file with content (just to simulate reading a file, though we use bytes directly)
	content := "this is the content of the test file"

	done := make(chan bool)
	var responseBody string
	var responseErr error

	// Read file content and send as binary data.
	fileContent := []byte(content)
	fetch.Post(baseURL+"/upload").
		Header("Content-Type", "application/octet-stream").
		Body(fileContent).
		Send(func(resp *fetch.Response, err error) {
			if err != nil {
				responseErr = err
			} else {
				responseBody = resp.Text()
			}
			done <- true
		})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error during file upload, got %v", responseErr)
	}
	if responseBody != content {
		t.Errorf("Expected echoed file content '%s', got '%s'", content, responseBody)
	}
}
