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
    // Create a fetchgo instance
    fg := fetchgo.New()
    
    // Create a client with base URL and timeout
    client := fg.NewClient("https://jsonplaceholder.typicode.com", 5000)

    // Send JSON request
    client.SendJSON("GET", "/posts/1", nil, func(result []byte, err error) {
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }

        fmt.Printf("Response: %s\n", string(result))
    })

    // Keep the program running to see the response
    select {}
}
```

## API Reference

### Core Types

#### `Fetchgo` struct

The main library struct that manages encoders and configuration.

```go
type Fetchgo struct {
    // Internal fields for TinyBin and CORS configuration
}

func New() *Fetchgo
```

#### `Client` interface

The HTTP client interface that provides methods for sending requests.

```go
type Client interface {
    SendJSON(method, url string, body any, callback func([]byte, error))
    SendBinary(method, url string, body any, callback func([]byte, error))
    SetHeader(key, value string)
}
```

#### `encoder` interface

Interface for encoding data, allowing pluggable serialization strategies.

```go
type encoder interface {
    Encode(data any) ([]byte, error)
}
```

### Client Methods

#### `SendJSON(method, url, body, callback)`

Sends an HTTP request with JSON encoding. The body is encoded as JSON and sent with `Content-Type: application/json; charset=utf-8`. The callback receives the raw response body as `[]byte`.

```go
func SendJSON(method, url string, body any, callback func([]byte, error))
```

#### `SendBinary(method, url, body, callback)`

Sends an HTTP request with TinyBin encoding. The body is encoded with TinyBin and sent with `Content-Type: application/octet-stream`. The callback receives the raw response body as `[]byte`.

```go
func SendBinary(method, url string, body any, callback func([]byte, error))
```

#### `SetHeader(key, value)`

Sets a default header that will be included in all requests from this client.

```go
func SetHeader(key, value string)
```

### Creating Clients

#### `NewClient(baseURL, timeoutMS)`

Creates a new HTTP client with the specified base URL and timeout.

```go
func (f *Fetchgo) NewClient(baseURL string, timeoutMS int) Client
```

**Parameters:**
- `method` - HTTP method (GET, POST, PUT, DELETE, etc.)
- `url` - Request URL (can be relative if baseURL is set, or absolute)
- `body` - Request body data to encode
- `callback` - Function called with response data as `[]byte` and error

**Examples:**

```go
// JSON request
client.SendJSON("POST", "/users", userData, func(result []byte, err error) {
    if err != nil {
        log.Printf("Request failed: %v", err)
        return
    }
    fmt.Printf("Response: %s", string(result))
})

// Binary request
client.SendBinary("POST", "/upload", fileData, func(result []byte, err error) {
    if err != nil {
        log.Printf("Upload failed: %v", err)
        return
    }
    fmt.Printf("Upload successful")
})
```

#### `SetHeader(key, value)`

Sets a default header that will be included in all requests from this client. Replaces any existing header with the same key.

```go
func SetHeader(key, value string)
```

**Example:**
```go
fg := fetchgo.New()
client := fg.NewClient("https://api.example.com", 5000)
client.SetHeader("Authorization", "Bearer token123")
client.SetHeader("Content-Type", "application/json")
```

### Data Encoding

#### JSON Encoding (`SendJSON`)

Automatically encodes request bodies as JSON with `Content-Type: application/json; charset=utf-8`. Response bodies are returned as raw `[]byte` for you to decode as needed.

#### TinyBin Encoding (`SendBinary`)

Automatically encodes request bodies using TinyBin serialization with `Content-Type: application/octet-stream`. Response bodies are returned as raw `[]byte` for you to decode as needed.

**Special case for raw bytes:** When sending `[]byte` data with `SendBinary`, the data is sent as-is without TinyBin encoding.

### Configuration

#### Base URL and Timeout

Configure the base URL and request timeout when creating a client:

```go
fg := fetchgo.New()
client := fg.NewClient("https://api.example.com", 5000) // 5 second timeout
```

#### Headers

Set default headers that apply to all requests:

```go
client.SetHeader("Authorization", "Bearer token123")
client.SetHeader("User-Agent", "MyApp/1.0")
```

### Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/cdvelop/fetchgo"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    fg := fetchgo.New()
    client := fg.NewClient("https://jsonplaceholder.typicode.com", 10000)
    
    client.SetHeader("Authorization", "Bearer mytoken")
    
    // Create a user
    user := User{Name: "John Doe", Email: "john@example.com"}
    
    client.SendJSON("POST", "/users", user, func(response []byte, err error) {
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        
        // Parse the JSON response
        var createdUser User
        if err := json.Unmarshal(response, &createdUser); err != nil {
            fmt.Printf("Failed to parse response: %v\n", err)
            return
        }
        
        fmt.Printf("Created user: %+v\n", createdUser)
    })
    
    // Keep the program running

## Platform Support

- **Server-side**: Uses Go's standard `net/http` package
- **Browser (WASM)**: Uses `syscall/js` to call JavaScript's fetch API
- **Cross-compilation**: Single codebase works across all platforms

## Dependencies

- `github.com/cdvelop/tinybin` - Binary serialization
- `github.com/cdvelop/tinystring` - String utility functions

## Migration from v1

If you're upgrading from the old API:

**Old API:**
```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    RequestType: fetchgo.RequestJSON,
}
client.SendRequest("POST", "/users", data, func(result any, err error) {
    // result was any, needed type assertion
})
```

**New API:**
```go
fg := fetchgo.New()
client := fg.NewClient("https://api.example.com", 5000)
client.SendJSON("POST", "/users", data, func(result []byte, err error) {
    // result is always []byte
})
```

## License

See LICENSE file for details.
