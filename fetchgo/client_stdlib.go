//go:build !wasm

package fetchgo

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/cdvelop/tinystring"
)

// doRequest is the standard library implementation for making an HTTP request.
func (c *Client) doRequest(method, endpoint string, body any) (any, error) {
	// 1. Build the full URL.
	fullURL, err := c.buildURL(endpoint)
	if err != nil {
		return nil, err
	}

	// 2. Prepare the request body.
	bodyReader, contentType, err := c.prepareBody(body)
	if err != nil {
		return nil, err
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
	if c.TimeoutMS > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(c.TimeoutMS)*time.Millisecond)
		defer cancel()
	}

	// 5. Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, Errf("failed to create request: %s", err.Error())
	}

	// Add headers to the request.
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 6. Execute the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, Errf("request failed: %s", err.Error())
	}
	defer resp.Body.Close()

	// 7. Read the response body.
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Errf("failed to read response body: %s", err.Error())
	}

	// 8. Check for non-successful status codes.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseBody, Errf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// prepareBody handles different body types for the standard library environment.
func (c *Client) prepareBody(body any) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
	}

	// Handle simple types that don't need an encoder.
	switch b := body.(type) {
	case []byte:
		return bytes.NewReader(b), "", nil // No default content type for raw bytes
	case string:
		// Assume string is a file path.
		file, err := os.Open(b)
		if err != nil {
			return nil, "", Errf("failed to open file %s: %s", b, err.Error())
		}
		// TODO: Add MIME type detection for files. For now, it's empty.
		return file, "", nil
	case io.Reader:
		return b, "", nil // Content type is unknown for a generic reader.
	}

	// For other types (structs, maps), use the configured encoder.
	encoderToUse := c.encoder
	var contentType string
	if encoderToUse == nil {
		switch c.RequestType {
		case RequestJSON:
			encoderToUse = &JSONEncoder{}
			contentType = "application/json; charset=utf-8"
		default: // Raw is the fallback
			encoderToUse = &RawEncoder{}
		}
	}

	encodedBody, err := encoderToUse.Encode(body)
	if err != nil {
		return nil, "", err
	}

	return bytes.NewReader(encodedBody), contentType, nil
}
