//go:build wasm

package fetch_test

import (
	"os"
	"testing"

	"github.com/tinywasm/fetch"
)

func TestWasm(t *testing.T) {
	// Read server URL from file (written by test server on startup)
	urlBytes, err := os.ReadFile(".test_server_url")
	if err != nil {
		t.Fatalf("Failed to read server URL from .test_server_url: %v. Make sure test server is running.", err)
	}
	serverURL := string(urlBytes)

	client := fetch.New().NewClient(serverURL, 0)

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, client) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, client) })
	t.Run("TimeoutSuccess", func(t *testing.T) {
		timeoutClient := fetch.New().NewClient(serverURL, 200)
		SendRequest_TimeoutSuccessShared(t, timeoutClient)
	})
	t.Run("TimeoutFailure", func(t *testing.T) {
		timeoutClient := fetch.New().NewClient(serverURL, 50)
		SendRequest_TimeoutFailureShared(t, timeoutClient)
	})
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, client) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, client) })
}
