package fetch

import (
	. "github.com/tinywasm/fmt"
)

// buildURL constructs the full request URL.
// Since the new API requires absolute URLs (or at least user-managed ones),
// this function now just validates the URL or passes it through.
// It is kept to satisfy existing calls in client_stdlib.go and client_wasm.go
func buildURL(url string) (string, error) {
	if url == "" {
		return "", Err("URL cannot be empty")
	}
	// In the previous implementation, we checked for absolute URLs if base was not set.
	// Now we don't have a base URL in the client, so we assume the user provides a valid URL.
	// We could enforce http/https prefix here if we wanted to be strict.

	// For now, let's just return it.
	return url, nil
}
