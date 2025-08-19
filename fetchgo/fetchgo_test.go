//go:build !wasm

package fetchgo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// setupTestServer creates a new httptest.Server with predefined handlers for testing.
func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Handler for simple GET requests
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get success"))
	})

	// Handler for JSON POST requests, echoes the body back
	mux.HandleFunc("/post_json", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Handler for file uploads
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Handler that simulates a slow response
	mux.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Handler that always returns an error status
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	return httptest.NewServer(mux)
}

func TestSendRequest_Get(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	done := make(chan bool)
	var responseBody []byte
	var responseErr error

	client.SendRequest("GET", "/get", nil, func(result any, err error) {
		if err != nil {
			responseErr = err
		} else if res, ok := result.([]byte); ok {
			responseBody = res
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

func TestSendRequest_PostJSON(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &Client{BaseURL: server.URL, RequestType: RequestJSON}
	done := make(chan bool)
	requestData := map[string]string{"message": "hello"}
	var responseBody []byte
	var responseErr error

	client.SendRequest("POST", "/post_json", requestData, func(result any, err error) {
		if err != nil {
			responseErr = err
		} else if res, ok := result.([]byte); ok {
			responseBody = res
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

func TestSendRequest_TimeoutSuccess(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &Client{BaseURL: server.URL, TimeoutMS: 200}
	done := make(chan bool)
	var responseErr error

	client.SendRequest("GET", "/timeout", nil, func(result any, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr != nil {
		t.Fatalf("Expected no error, but request timed out: %v", responseErr)
	}
}

func TestSendRequest_TimeoutFailure(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &Client{BaseURL: server.URL, TimeoutMS: 50}
	done := make(chan bool)
	var responseErr error

	client.SendRequest("GET", "/timeout", nil, func(result any, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr == nil {
		t.Fatal("Expected request to time out, but it succeeded.")
	}
}

func TestSendRequest_ServerError(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	done := make(chan bool)
	var responseErr error

	client.SendRequest("GET", "/error", nil, func(result any, err error) {
		responseErr = err
		done <- true
	})

	<-done

	if responseErr == nil {
		t.Fatal("Expected an error for 500 status code, but got nil.")
	}
}

func TestSendRequest_PostFile(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

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

	client := &Client{BaseURL: server.URL}
	done := make(chan bool)
	var responseBody []byte
	var responseErr error

	// Send the file path as the body.
	client.SendRequest("POST", "/upload", tmpfile.Name(), func(result any, err error) {
		if err != nil {
			responseErr = err
		} else if res, ok := result.([]byte); ok {
			responseBody = res
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
