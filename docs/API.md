# API Reference

## Core Types

### `Fetch` struct

The main library struct that manages encoders and configuration.

```go
type Fetch struct {
    // Internal fields for TinyBin and CORS configuration
}

func New() *Fetch
```

### `Client` interface

The HTTP client interface that provides methods for sending requests.

```go
type Client interface {
    SendJSON(method, url string, body any, callback func([]byte, error))
    SendBinary(method, url string, body any, callback func([]byte, error))
    SetHeader(key, value string)
}
```

### `encoder` interface

Interface for encoding data, allowing pluggable serialization strategies.

```go
type encoder interface {
    Encode(data any) ([]byte, error)
}
```

## Client Methods

### `SendJSON(method, url, body, callback)`

Sends an HTTP request with JSON encoding. The body is encoded as JSON and sent with `Content-Type: application/json; charset=utf-8`. The callback receives the raw response body as `[]byte`.

```go
func SendJSON(method, url string, body any, callback func([]byte, error))
```

### `SendBinary(method, url, body, callback)`

Sends an HTTP request with TinyBin encoding. The body is encoded with TinyBin and sent with `Content-Type: application/octet-stream`. The callback receives the raw response body as `[]byte`.

```go
func SendBinary(method, url string, body any, callback func([]byte, error))
```

**Special case for raw bytes:** When sending `[]byte` data with `SendBinary`, the data is sent as-is without TinyBin encoding.

### `SetHeader(key, value)`

Sets a default header that will be included in all requests from this client. Replaces any existing header with the same key.

```go
func SetHeader(key, value string)
```

## Creating Clients

### `NewClient(baseURL, timeoutMS)`

Creates a new HTTP client with the specified base URL and timeout.

```go
func (f *Fetch) NewClient(baseURL string, timeoutMS int) Client
```

## Configuration

### Base URL and Timeout

Configure the base URL and request timeout when creating a client:

```go
fg := fetchgo.New()
client := fg.NewClient("https://api.example.com", 5000) // 5 second timeout
```

### Headers

Set default headers that apply to all requests:

```go
client.SetHeader("Authorization", "Bearer token123")
client.SetHeader("User-Agent", "MyApp/1.0")
```
