//go:build !wasm

package fetchgo_test

import (
	"testing"

	"github.com/cdvelop/fetchgo"
)

func TestStdlib(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	client := &fetchgo.Client{BaseURL: server.URL}

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, client) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, client) })
	t.Run("TimeoutSuccess", func(t *testing.T) { SendRequest_TimeoutSuccessShared(t, client) })
	t.Run("TimeoutFailure", func(t *testing.T) { SendRequest_TimeoutFailureShared(t, client) })
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, client) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, client) })
}
