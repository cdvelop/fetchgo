//go:build !wasm

package fetch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	. "github.com/tinywasm/fmt"
)

// doRequest is the standard library implementation for making an HTTP request.
func (c *client) doRequest(method, endpoint string, contentType string, body []byte, callback func([]byte, error)) {
	go func() {
		// 1. Build the full URL.
		fullURL, err := c.buildURL(endpoint)
		if err != nil {
			callback(nil, err)
			return
		}

		// 2. Prepare body reader.
		var bodyReader io.Reader
		if len(body) > 0 {
			bodyReader = bytes.NewReader(body)
		}

		// 3. Prepare the headers.
		headers := c.getHeaders()
		if contentType != "" {
			if headers == nil {
				headers = make(map[string]string)
			}
			if _, exists := headers["Content-Type"]; !exists {
				headers["Content-Type"] = contentType
			}
		}

		// 4. Set up the request context with timeout.
		ctx := context.Background()
		if c.timeoutMS > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(c.timeoutMS)*time.Millisecond)
			defer cancel()
		}

		// 5. Create the HTTP request.
		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			callback(nil, Errf("failed to create request: %s", err.Error()))
			return
		}

		// Add headers to the request.
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// 6. Execute the request.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			callback(nil, Errf("request failed: %s", err.Error()))
			return
		}
		defer resp.Body.Close()

		// 7. Read the response body.
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			callback(nil, Errf("failed to read response body: %s", err.Error()))
			return
		}

		// 8. Check for non-successful status codes.
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			callback(responseBody, Errf("request failed with status %d: %s", resp.StatusCode, string(responseBody)))
			return
		}

		callback(responseBody, nil)
	}()
}
