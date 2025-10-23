# fetchgo

fetchgo is a minimal Go library designed to simplify HTTP requests across different runtimes. It provides a unified API that works in the browser with TinyGo + WebAssembly using syscall/js, and on the server using Go's standard net/http package. With fetchgo, you can write cross-platform HTTP logic once and run it anywhere.

## Installation

```bash
go get github.com/cdvelop/fetchgo
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/cdvelop/fetchgo"
)

func main() {
    client := &fetchgo.Client{
        BaseURL: "https://jsonplaceholder.typicode.com",
    }

    client.SendRequest("GET", "/posts/1", nil, func(result any, err error) {
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }

        if body, ok := result.([]byte); ok {
            fmt.Printf("Response: %s\n", string(body))
        }
    })

    // Keep the program running to see the response
    select {}
}
```

## API Reference

### Core Types

#### `Fetchgo` struct

The main library struct. Currently used primarily for initialization.

```go
type Fetchgo struct{}

func New() *Fetchgo
```

#### `Client` struct

The main HTTP client that handles requests across different platforms.

```go
type Client struct {
    BaseURL        string      // Base URL for all requests, e.g., "https://api.example.com"
    defaultHeaders []string    // Internal storage for headers: ["key1", "value1", "key2", "value2"]
    TimeoutMS      int         // Request timeout in milliseconds
    RequestType    requestType // Default request type (e.g., RequestJSON, RequestRaw)
    encoder        encoder     // Optional: custom encoder for request bodies
}
```

#### `encoder` interface

Interface for encoding and decoding data, allowing pluggable serialization strategies.

```go
type encoder interface {
    Encode(data any) ([]byte, error)
    Decode(data []byte, v any) error
}
```

### Request Types

Constants that define how request bodies should be encoded:

- `RequestJSON` - Encode request body as JSON (`application/json`)
- `RequestForm` - Encode request body as form data (`application/x-www-form-urlencoded`)
- `RequestMultipart` - Encode request body as multipart form data (`multipart/form-data`)
- `RequestRaw` - Send request body as-is (raw bytes)

### Client Methods

#### `SendRequest(method, url, body, callback)`

Sends an HTTP request asynchronously and invokes the callback with the result.

```go
func (c *Client) SendRequest(method, url string, body any, callback func(any, error))
```

**Parameters:**
- `method` - HTTP method (GET, POST, PUT, DELETE, etc.)
- `url` - Request URL (can be relative if BaseURL is set, or absolute)
- `body` - Request body (type depends on RequestType or custom encoder)
- `callback` - Function called with response data and error

**Example:**
```go
client.SendRequest("POST", "/users", userData, func(result any, err error) {
    if err != nil {
        log.Printf("Request failed: %v", err)
        return
    }

    if body, ok := result.([]byte); ok {
        fmt.Printf("Response: %s", string(body))
    }
})
```

#### `AddHeader(key, value)`

Adds a header to the client's default headers. Allows duplicate header keys.

```go
func (c *Client) AddHeader(key, value string)
```

**Example:**
```go
client := &fetchgo.Client{BaseURL: "https://api.example.com"}
client.AddHeader("Authorization", "Bearer token123")
client.AddHeader("X-Custom-Header", "value")
```

#### `SetHeader(key, value)`

Sets a header, ensuring there's at most one entry for the given key. Replaces existing values.

```go
func (c *Client) SetHeader(key, value string)
```

**Example:**
```go
client := &fetchgo.Client{BaseURL: "https://api.example.com"}
client.SetHeader("Authorization", "Bearer token123")
client.SetHeader("Authorization", "Bearer newtoken456") // Replaces previous value
```

### Built-in Encoders

#### `JSONEncoder`

Implements JSON encoding/decoding for request and response bodies.

```go
type JSONEncoder struct{}

func (e JSONEncoder) Encode(data any) ([]byte, error)
func (e JSONEncoder) Decode(data []byte, v any) error
```

**Example:**
```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    RequestType: fetchgo.RequestJSON,
}

// Send JSON data
data := map[string]string{"name": "John", "email": "john@example.com"}
client.SendRequest("POST", "/users", data, func(result any, err error) {
    // result will be []byte containing JSON response
})
```

#### `RawEncoder`

Implements raw byte encoding/decoding. Useful for sending files or binary data.

```go
type RawEncoder struct{}

func (e RawEncoder) Encode(data any) ([]byte, error)
func (e RawEncoder) Decode(data []byte, v any) error
```

**Example:**
```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    RequestType: fetchgo.RequestRaw,
}

// Send raw bytes
rawData := []byte("raw binary data")
client.SendRequest("POST", "/upload", rawData, func(result any, err error) {
    // result will be []byte containing raw response
})

// Send file path (will be read as bytes)
client.SendRequest("POST", "/upload", "/path/to/file.txt", func(result any, err error) {
    // result will be []byte containing file contents
})
```

### Configuration Options

#### BaseURL

Set the base URL for all requests. Relative URLs will be resolved against this base.

```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
}

// These are equivalent:
// client.SendRequest("GET", "/users", nil, callback)
// client.SendRequest("GET", "https://api.example.com/users", nil, callback)
```

#### TimeoutMS

Set request timeout in milliseconds.

```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    TimeoutMS: 5000, // 5 second timeout
}
```

#### RequestType

Set the default request type for all requests.

```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    RequestType: fetchgo.RequestJSON,
}
```

#### Custom encoder

Use a custom encoder for specialized serialization needs.

```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    encoder: &CustomEncoder{},
}
```

### Advanced Usage Examples

#### Multiple Headers

```go
client := &fetchgo.Client{BaseURL: "https://api.example.com"}

// Add multiple headers
client.AddHeader("Authorization", "Bearer token123")
client.AddHeader("Content-Type", "application/json")
client.AddHeader("X-Client-Version", "1.0")

client.SendRequest("GET", "/data", nil, func(result any, err error) {
    // All headers will be sent with the request
})
```

#### Different Request Types

```go
client := &fetchgo.Client{BaseURL: "https://httpbin.org"}

// JSON request
jsonData := map[string]string{"name": "John"}
client.SendRequest("POST", "/post", jsonData, func(result any, err error) {
    // Content-Type: application/json
})

// Raw request
rawData := []byte("raw data")
client.SendRequest("POST", "/post", rawData, func(result any, err error) {
    // Content-Type: application/octet-stream
})
```

#### Error Handling

```go
client := &fetchgo.Client{BaseURL: "https://api.example.com"}

client.SendRequest("GET", "/users/999", nil, func(result any, err error) {
    if err != nil {
        // Handle different types of errors
        log.Printf("Request failed: %v", err)
        return
    }

    // Process successful response
    if body, ok := result.([]byte); ok {
        fmt.Printf("Response: %s", string(body))
    }
})
```

## Platform Support

- **Server-side**: Uses Go's standard `net/http` package
- **Browser (WASM)**: Uses `syscall/js` to call JavaScript's fetch API
- **Cross-compilation**: Single codebase works across all platforms

## Dependencies

- `github.com/cdvelop/tinystring` - String utility functions

## License

See LICENSE file for details.
