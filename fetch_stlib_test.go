//go:build !wasm

package fetch_test

import (
	"testing"
)

func TestStdlib(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, server.URL) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, server.URL) })
	t.Run("TimeoutSuccess", func(t *testing.T) { SendRequest_TimeoutSuccessShared(t, server.URL) })
	t.Run("TimeoutFailure", func(t *testing.T) { SendRequest_TimeoutFailureShared(t, server.URL) })
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, server.URL) })
	t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, server.URL) })
}
