//go:build wasm

package fetchgo_test

import (
	"testing"

	"github.com/cdvelop/fetchgo"
)

func TestWasm(t *testing.T) {
	// Note: For WASM tests, a test server should be running on localhost:8080
	// You can start it manually or modify the test setup accordingly.
	client := &fetchgo.Client{BaseURL: "http://localhost:8080"}

	t.Run("Get", func(t *testing.T) { SendRequest_GetShared(t, client) })
	t.Run("PostJSON", func(t *testing.T) { SendRequest_PostJSONShared(t, client) })
	t.Run("TimeoutSuccess", func(t *testing.T) { SendRequest_TimeoutSuccessShared(t, client) })
	t.Run("TimeoutFailure", func(t *testing.T) { SendRequest_TimeoutFailureShared(t, client) })
	t.Run("ServerError", func(t *testing.T) { SendRequest_ServerErrorShared(t, client) })
	// PostFile test skipped in WASM - requires actual File object from browser input
	// In WASM, files must come from <input type="file"> or drag-and-drop, not filesystem paths
	// t.Run("PostFile", func(t *testing.T) { SendRequest_PostFileShared(t, client) })
}
