//go:build !wasm

package fetchgo_test

import (
	"testing"

	"github.com/cdvelop/fetchgo"
)

func TestStdlib(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := fetchgo.New().NewClient(server.URL, 0)

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, client) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, client) })
	t.Run("TimeoutSuccess", func(t *testing.T) {
		timeoutClient := fetchgo.New().NewClient(server.URL, 200)
		SendRequest_TimeoutSuccessShared(t, timeoutClient)
	})
	t.Run("TimeoutFailure", func(t *testing.T) {
		timeoutClient := fetchgo.New().NewClient(server.URL, 50)
		SendRequest_TimeoutFailureShared(t, timeoutClient)
	})
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, client) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, client) })
}
