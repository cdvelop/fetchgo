//go:build !wasm

package fetch_test

import (
	"testing"

	"github.com/tinywasm/fetch"
)

func TestStdlib(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := fetch.New().NewClient(server.URL, 0)

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, client) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, client) })
	t.Run("TimeoutSuccess", func(t *testing.T) {
		timeoutClient := fetch.New().NewClient(server.URL, 200)
		SendRequest_TimeoutSuccessShared(t, timeoutClient)
	})
	t.Run("TimeoutFailure", func(t *testing.T) {
		timeoutClient := fetch.New().NewClient(server.URL, 50)
		SendRequest_TimeoutFailureShared(t, timeoutClient)
	})
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, client) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, client) })
}
