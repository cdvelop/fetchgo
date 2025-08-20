package fetchgo

// Client is a configurable HTTP client that works across WASM and standard Go.
// To create a client, initialize the struct directly, for example:
// client := &fetchgo.Client{ BaseURL: "https://api.example.com", RequestType: fetchgo.RequestJSON }
type Client struct {
	BaseURL        string      // Base URL for all requests, e.g., "https://api.example.com"
	defaultHeaders []string    // Internal storage for headers: ["key1", "value1", "key2", "value2"]
	TimeoutMS      int         // Request timeout in milliseconds.
	RequestType    requestType // Default request type (e.g., RequestJSON, RequestRaw).
	encoder        Encoder     // Optional: custom encoder for request bodies. If nil, a default is used based on RequestType.
}

// SendRequest sends an HTTP request and invokes the callback with the result.
// It delegates the actual request logic to the environment-specific doRequest method.
// The entire operation, including the callback, runs in a new goroutine.
func (c *Client) SendRequest(method, url string, body any, callback func(any, error)) {
	go func() {
		// doRequest is implemented in client_stdlib.go and client_wasm.go.
		// It handles URL building, body preparation, and the HTTP request itself.
		result, err := c.doRequest(method, url, body)
		callback(result, err)
	}()
}
