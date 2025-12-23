package fetch

import (
	"strings"
)

// Header represents a single HTTP header key-value pair.
type Header struct {
	Key   string
	Value string
}

// Request represents an HTTP request builder.
type Request struct {
	method  string
	url     string
	headers []Header
	body    []byte
	timeout int
}

// Response represents an HTTP response.
type Response struct {
	Status     int
	Headers    []Header
	RequestURL string
	Method     string
	body       []byte
}

// Get creates a new GET request.
func Get(url string) *Request {
	return &Request{method: "GET", url: url}
}

// Post creates a new POST request.
func Post(url string) *Request {
	return &Request{method: "POST", url: url}
}

// Put creates a new PUT request.
func Put(url string) *Request {
	return &Request{method: "PUT", url: url}
}

// Delete creates a new DELETE request.
func Delete(url string) *Request {
	return &Request{method: "DELETE", url: url}
}

// Header adds a header to the request.
func (r *Request) Header(key, value string) *Request {
	r.headers = append(r.headers, Header{Key: key, Value: value})
	return r
}

// Body sets the request body.
func (r *Request) Body(data []byte) *Request {
	r.body = data
	return r
}

// Timeout sets the request timeout in milliseconds.
func (r *Request) Timeout(ms int) *Request {
	r.timeout = ms
	return r
}

// Send executes the request and calls the callback with the response.
func (r *Request) Send(callback func(*Response, error)) {
	doRequest(r, callback)
}

// Dispatch executes the request and sends the response to the global handler.
// This is a fire-and-forget method.
func (r *Request) Dispatch() {
	if globalHandler == nil {
		log("Dispatch called but no global handler set")
		return
	}
	doRequest(r, func(resp *Response, err error) {
		if err != nil {
			log("Dispatch error:", err)
			return
		}
		globalHandler(resp)
	})
}

// Body returns the response body as a byte slice.
func (r *Response) Body() []byte {
	return r.body
}

// Text returns the response body as a string.
func (r *Response) Text() string {
	return string(r.body)
}

// GetHeader returns the value of the specified header.
// It is case-insensitive.
func (r *Response) GetHeader(key string) string {
	key = strings.ToLower(key)
	for _, h := range r.Headers {
		if strings.ToLower(h.Key) == key {
			return h.Value
		}
	}
	return ""
}
