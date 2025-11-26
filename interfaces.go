package fetchgo

// Client defines the public interface for an HTTP client.
type Client interface {
	// SendJSON performs an HTTP request with JSON encoding.
	// url MUST be absolute (e.g., "https://api.example.com/users")
	// body is encoded to JSON, Content-Type: application/json
	// The callback receives the raw response body as []byte.
	// The user is responsible for decoding it using json.Unmarshal.
	SendJSON(method, url string, body any, callback func([]byte, error))

	// SendBinary performs an HTTP request with TinyBin encoding.
	// url MUST be absolute (e.g., "https://api.example.com/users")
	// body is encoded with TinyBin, Content-Type: application/octet-stream
	// The callback receives the raw response body as []byte.
	// The user is responsible for decoding it using tinybin.Decode.
	SendBinary(method, url string, body any, callback func([]byte, error))

	// SetHeader sets a default header for all requests from this client.
	SetHeader(key, value string)
}
