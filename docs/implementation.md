# FetchGo Library Improvement Proposal

## Overview

This document outlines the proposed unification of the existing `fetchclient` and `fetchserver` modules into a single, unified `fetchgo` library. The goal is to create a seamless HTTP client API that works across different compilation targets using Go build tags.

> Note: The deprecated libraries marked for removal after implementation are located in `deprecatedCode/fetchclient/` and `deprecatedCode/fetchserver/`.

## Current State Analysis

### FetchClient Module
- **Target Environment**: WebAssembly (WASM) in browser
- **Dependencies**: `syscall/js` for browser API access
- **Main Functionality**:
  - JavaScript Fetch API integration
  - Promise handling with callbacks
  - AbortController support for request cancellation
  - FormData and JSON request body support
  - Response handling with automatic JSON parsing

### FetchServer Module  
- **Target Environment**: Standard Go runtime (server/desktop)
- **Dependencies**: `net/http` standard library
- **Main Functionality**:
  - HTTP client using Go's standard library
  - Support for GET/POST methods
  - JSON and multipart/form-data request bodies
  - File download capability
  - URL parameter encoding

### Current Interface (from model package)
Both modules implement a common interface through the `model.MainHandler`:
```go
type FetchAdapter interface {
    SendOneRequest(method, endpoint, object string, body_rq any, response func(result []map[string]string, err string))
    SendAllRequests(endpoint string, data []model.Response, response func(result []model.Response, err string))
}
```

## Proposed Architecture

### Build Tag Strategy
- **WASM Build**: `//go:build wasm`
- **Standard Library Build**: `//go:build !wasm`


## Decisions Made

### 1. API Design Philosophy ✅ DECIDED
**Decision**: Asynchronous callback-based API compatible with TinyGo

**Rationale**: 
- TinyGo compatibility is critical
- Callbacks work consistently across WASM and standard Go
- Simpler to implement than channels for cross-platform async

**Final API**: `SendRequest(method, url string, body any, callback func(data any, err error))`

### 2. Request/Response Types ✅ DECIDED
**Decision**: Flexible `any` type approach

**Rationale**:
- No validation at library level - responsibility of the caller
- Maximum flexibility for different use cases
- Simpler implementation without complex type constraints

**API**: 
- Input: `any` (caller responsibility to provide valid data)
- Output: `any` (caller responsibility to handle/validate response)

### 3. Error Handling ✅ DECIDED
**Decision**: Unified error handling with standard Go error interface

**Implementation**:
- Single error type across both environments
- WASM errors converted to standard Go errors
- Consistent error handling in callbacks

### 4. Dependencies ✅ DECIDED
**Decision**: Remove `github.com/cdvelop/model` dependency

**Rationale**:
- Make library self-contained
- Reduce external dependencies
- Improve maintainability and adoption

## Final Decisions Made

### 5. Async Implementation Strategy ✅ DECIDED
**Decision**: **Goroutines + Callbacks**

**Implementation**:
```go
func (c *Client) SendRequest(method, url string, body any, callback func(any, error)) {
    go func() {
        result, err := c.doRequest(method, url, body)
        callback(result, err)
    }()
}
```

### 6. Data Encoding/Decoding Strategy ✅ DECIDED
**Decision**: **Pluggable Encoder Interface**

**Implementation**:
```go
type Encoder interface {
    Encode(any) ([]byte, error)
    Decode([]byte, &any) error
}

type Client struct {
    encoder Encoder // Optional, defaults NONE
}
```

### 7. Request Configuration ✅ DECIDED  
**Decision**: **Simple Configuration with TinyGo optimizations**

**Implementation**:
```go
// requestType is a private type that defines how the request body should be
// encoded and which Content-Type header to set. The constants are exported
// so callers can select request kinds without depending on a public type.
type requestType string

const (
    RequestJSON     requestType = "json"
    RequestForm     requestType = "form"
    RequestMultipart requestType = "multipart"
    RequestRaw      requestType = "raw"
)

type Client struct {
    BaseURL        string // e.g. "https://api.example.com"
    defaultHeaders []string    // TinyGo friendly: ["key1", "value1", "key2", "value2"] (private)
    TimeoutMS int              // Numeric timeout in milliseconds, environment-specific handling
    RequestType    requestType // private: use Client options or helpers to set
    encoder        Encoder
}
```

### 8. Error Handling ✅ DECIDED
**Decision**: **Simple errors using tinystring library**

**Implementation**:
```go
// Use tinystring.Err() and tinystring.Errf() instead of standard library
// Use tinystring.Conv for string/number conversions
import . "github.com/cdvelop/tinystring"

// Example error creation:
err := Errf("fetch failed: status %s", statusCode)
```

### 9. Testing Strategy ✅ DECIDED
**Decision**: **Standard library testing now, Playwright-go for WASM later**

**Current Phase**: Focus on standard library implementation testing
**Future**: Playwright-go integration for WASM testing

## Remaining Questions and Considerations

### 1. 🔧 **TinyGo Header Handling Strategy**
**Question**: How to efficiently handle the `[]string` headers format?

**Proposed Implementation**:
```go
type Client struct {
    DefaultHeaders []string // ["Content-Type", "application/json", "Authorization", "Bearer token"]
}

// Internal helper that implements the core logic for adding/replacing
// headers. If replace==true it will replace the first occurrence of key,
// otherwise it will append a new (key,value) pair even if the key exists.
func (c *Client) upsertHeader(key, value string, replace bool) {
    for i := 0; i < len(c.defaultHeaders); i += 2 {
        if i+1 < len(c.defaultHeaders) && c.defaultHeaders[i] == key {
            if replace {
                c.defaultHeaders[i+1] = value
                return
            }
            // add duplicate and return
            c.defaultHeaders = append(c.defaultHeaders, key, value)
            return
        }
    }
    // not found: append
    c.defaultHeaders = append(c.defaultHeaders, key, value)
}

// Public helper to add headers. Adds the (key,value) pair even if the key
// already exists (may create duplicates).
func (c *Client) AddHeader(key, value string) {
    c.upsertHeader(key, value, false)
}

// SetHeader ensures there is at most one entry for the given key. If the
// header already exists it is replaced; otherwise the header is appended.
func (c *Client) SetHeader(key, value string) {
    c.upsertHeader(key, value, true)
}

// getHeaders converts the private []string representation to a map for
// internal use. If duplicate keys exist the last value wins (consistent
// with SetHeader behavior and overwrite semantics).
func (c *Client) getHeaders() map[string]string {
    headers := make(map[string]string)
    for i := 0; i < len(c.defaultHeaders); i += 2 {
        if i+1 < len(c.defaultHeaders) {
            headers[c.defaultHeaders[i]] = c.defaultHeaders[i+1]
        }
    }
    return headers
}
```

**Questions**:
- Should we provide helper methods for header manipulation?
- How to handle header validation (odd number of elements)?
- Should we panic or return error on malformed headers?

### 2. ⏱️ **Environment-Specific Timeout Handling**
**Question**: How to implement timeout differently for WASM vs StdLib?

**Options**:

**A) Build-tag specific timeout implementation**
```go
// client_wasm.go
func (c *Client) applyTimeout(timeoutMS int) js.Value {
    // Use AbortController with setTimeout (milliseconds)
    controller := js.Global().Get("AbortController").New()
    if timeoutMS > 0 {
        js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
            controller.Call("abort")
            return nil
        }), timeoutMS)
    }
    return controller
}

// client_stdlib.go  
func (c *Client) applyTimeout(timeoutMS int) context.Context {
    if timeoutMS > 0 {
        ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeoutMS)*time.Millisecond)
        return ctx
    }
    return context.Background()
}
```

**B) Unified timeout interface**
```go
type TimeoutHandler interface {
    ApplyTimeout(seconds int) // Environment-specific implementation
    Cancel()                  // Cancel current timeout
}
```

**My Recommendation**: **Option A** - Direct build-tag implementation for simplicity

### 3. 📝 **Request Type Handling**
**Question**: How should `RequestType` field influence request processing?

**Proposed Behavior**:
```go
// Impact on Content-Type header and body encoding
switch c.requestType {
case requestTypeJSON:
    // Content-Type: application/json
    // Use JSON encoder
case requestTypeForm:
    // Content-Type: application/x-www-form-urlencoded  
    // Use URL encoding
case requestTypeMultipart:
    // Content-Type: multipart/form-data
    // Use multipart encoder
case requestTypeRaw:
    // No Content-Type set
    // Pass body as-is
}
```

**Questions**:
- Should RequestType be per-client or per-request?
- How to override RequestType for specific requests?
- Should we support custom Content-Type headers?

### 4. 🔌 **Encoder Interface Implementation**
**Question**: What default encoders should we provide?

**Proposed Default Encoders**:
```go
// JSONEncoder using encoding/json
type JSONEncoder struct{}
func (e JSONEncoder) Encode(data any) ([]byte, error) { ... }
func (e JSONEncoder) Decode(data []byte) (any, error) { ... }

// FormEncoder for URL-encoded forms  
type FormEncoder struct{}
func (e FormEncoder) Encode(data any) ([]byte, error) { ... }
func (e FormEncoder) Decode(data []byte) (any, error) { ... }

// RawEncoder for []byte pass-through
type RawEncoder struct{}
func (e RawEncoder) Encode(data any) ([]byte, error) { ... }
func (e RawEncoder) Decode(data []byte) (any, error) { ... }
```

**Questions**:
- Should we include all encoders or just JSON initially?
- How to handle encoder selection based on RequestType?
- Should encoders handle type assertions internally?

### 5. � File and Binary Payload Handling

Decision: Remove a separate UploadFile API. The single `SendRequest` entrypoint will accept files and binary payloads directly. This keeps the surface area small and lets callers provide the body in the most natural form for their environment.

Guidelines:

- Acceptable `body` values for `SendRequest`:
    - WASM: `js.Value` (File object or other JS body) — pass through to the browser Fetch API when appropriate
    - StdLib: `[]byte`, `string` (file path), or `io.Reader`

- Internal helper idea (implementation detail):
```go
// prepareBody inspects `body` and returns a representation suitable for
// the environment-specific request executor.
// For StdLib it returns ([]byte, contentType string, io.ReadCloser, error).
// For WASM it may return (js.Value, contentType string, error) so the
// Fetch call can use a File/Blob directly without extra copies.
func (c *Client) prepareBody(body any) (interface{}, string, error) {
        // detect types and prepare body accordingly
}
```

- Notes:
    - For WASM, if `body` is a `js.Value` File/Blob, prefer passing it directly to `fetch` to avoid copying large files into Go memory.
    - For StdLib, accept file paths and `io.Reader` and stream them into the request body when possible.
    - MIME detection can be attempted (e.g., via file extension or magic bytes) but is optional; callers can always set `Content-Type` via headers.
    - Progress reporting and advanced streaming for very large uploads are considered Phase 2 features.

This change replaces the previously separate Upload API and consolidates file and binary handling into `SendRequest`, keeping the public API surface smaller and easier to migrate to.

### 6. 🎯 **SendRequest Method Versatility**
**Question**: How to make SendRequest handle any type of request efficiently?

**Proposed Signature**:
```go
func (c *Client) SendRequest(method, url string, body any, callback func(any, error)) {
    // Method can be: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
    // URL can be relative (uses BaseURL) or absolute
    // Body can be: nil, []byte, string, struct, map, js.Value (for WASM files)
    // Callback receives: response data (any) and error
}
```

**Internal Processing Logic**:
```go
func (c *Client) SendRequest(method, url string, body any, callback func(any, error)) {
    go func() {
        // 1. Build full URL (BaseURL + url if relative)
        fullURL := c.buildURL(url)
        
        // 2. Determine encoding based on RequestType and body type
        encodedBody, contentType, err := c.encodeBody(body)
        if err != nil {
            callback(nil, tinystring.Errf("encoding error: %s", err.Error()))
            return
        }
        
        // 3. Build headers (DefaultHeaders + Content-Type)
        headers := c.buildHeaders(contentType)
        
        // 4. Execute request (environment-specific)
        result, err := c.doRequest(method, fullURL, encodedBody, headers)
        
        // 5. Process response
        callback(result, err)
    }()
}
```

**Questions**:
- Should we validate HTTP methods?
- How to handle special cases like HEAD requests (no body expected in response)?
- Should we auto-retry on certain errors?

### 7. 📦 **Dependency Management**
**Question**: How to integrate tinystring dependency?

**Required from tinystring**:
```go
import . "github.com/cdvelop/tinystring"

// Error handling
err := Err("simple error message")
err := Errf("formatted error: %s", details)

// String/number conversions
statusStr := Convert(statusCode).String()
timeout := Convert(timeoutStr).Int()
```

**Questions**:
- Should we wrap tinystring functions for internal use?
- How to handle tinystring availability in different environments?
- Should we have fallback implementations if tinystring is not available?

### 8. 🗂️ **File Structure Refinement**
**Question**: Final file organization based on decisions?

**Proposed Structure**:
```
fetchgo/
├── client.go              // Main Client struct and public API
├── types.go              // Interfaces (Encoder) and types
├── encoders.go           // Default encoder implementations  
├── client_wasm.go        // WASM-specific implementation (//go:build wasm)
├── client_stdlib.go      // StdLib implementation (//go:build !wasm)
├── utils.go              // Shared utilities (URL building, header processing)
├── timeout_wasm.go       // WASM timeout handling (//go:build wasm)
├── timeout_stdlib.go     // StdLib timeout handling (//go:build !wasm)
└── upload.go             // File upload utilities
```

**Questions**:
- Should timeout handling be in separate files or integrated?
- Do we need separate upload files or integrate in client files?
- Should encoders be in the same package or separate?

## Proposed Implementation Plan

### Phase 1: Core Structure
1. Define unified interfaces and types
2. Implement basic client structure
3. Create build tag separated implementations
4. Basic request/response handling

### Phase 2: Feature Parity
1. Port existing functionality from both modules
2. Implement unified error handling
3. Add request cancellation support
4. File upload/download capabilities

### Phase 3: Enhancements
1. Add middleware support
2. Implement configuration options
3. Add comprehensive testing
4. Documentation and examples

### Phase 4: Migration Support
1. Provide migration guide from old modules
2. Compatibility layer (if needed)
3. Deprecation timeline for old modules

### Dependencies Analysis

### Current Dependencies (TO BE REMOVED)
- **fetchclient**: `github.com/cdvelop/model` ❌ 
- **fetchserver**: `github.com/cdvelop/model` ❌

### New Dependencies (Minimal)
- **Required**: `github.com/cdvelop/tinystring` for error handling and conversions
- **Standard Library**: `encoding/json`, `net/http`, `syscall/js` (WASM only)
- **Zero External Dependencies**: Besides tinystring, completely self-contained

### Functions to Replace from model Package
1. **DataConverter.EncodeMaps()** → Pluggable Encoder interface
2. **DataConverter.DecodeMaps()** → Pluggable Encoder interface  
3. **Logger interface** → Removed (keep library simple)
4. **Error handling** → `tinystring.Err()` and `tinystring.Errf()`

## Breaking Changes Considerations

### Potential Breaking Changes
1. Package name change from `fetchclient`/`fetchserver` to `fetchgo`
2. API signature changes (if we move away from callbacks)
3. Error handling changes
4. Import path changes

### Migration Strategy
1. Provide clear migration documentation
2. Consider compatibility packages
3. Version strategy for smooth transition

## Final Questions Requiring Decision

### 1. � **Header Helper Methods**
**Question**: Should we provide helper methods for header manipulation?

**Options**:
- **A)** Full helper API: `AddHeader()`, `RemoveHeader()`, `GetHeaders()`
- **B)** Minimal helpers: Just `AddHeader()` 
- **C)** No helpers: Direct []string manipulation

**My Recommendation**: **Option B** - AddHeader() for convenience, direct access for advanced use

### 2. ⏱️ **Timeout Implementation**
**Question**: Approve the build-tag specific timeout approach?

**Proposed**: AbortController + setTimeout for WASM, context.WithTimeout for StdLib
**Your approval needed**: Is this approach acceptable?

### 3. 🎯 **RequestType Scope** 
**Question**: Should RequestType be per-client or allow per-request override?

**Options**:
- **A)** Per-client only (simpler)
- **B)** Per-client with per-request override parameter
- **C)** Per-request only (more complex API)

**My Recommendation**: **Option A** - Keep it simple initially

### 4. 🔌 **Default Encoders**
**Question**: Which encoders to include initially?

**Options**:
- **A)** JSON only (minimal)
- **B)** JSON + Form + Raw (comprehensive)  
- **C)** JSON + Raw (balanced)

**My Recommendation**: **Option C** - Covers most use cases without complexity

### 5. � **File Upload Priority**
**Question**: Should file upload be implemented in Phase 1 or Phase 2?

**Consideration**: File upload adds complexity but is mentioned as important
**Your decision needed**: Priority level for file upload feature?

### 6. 🧪 **Testing Approach Confirmation**
**Question**: Start with standard library testing only?

**Proposed**: Focus on StdLib implementation testing first, WASM testing later with Playwright-go
**Your confirmation needed**: Is this acceptable for initial development?

## Implementation Roadmap

### Phase 1: Core Foundation (Immediate)
1. ✅ **Create project structure** with build tags
2. ✅ **Implement Client struct** with TinyGo-optimized fields
3. ✅ **Create Encoder interface** and JSON/Raw implementations
4. ✅ **Build basic SendRequest** with goroutines + callbacks
5. ✅ **Implement environment-specific timeout handling**
6. ✅ **Add tinystring integration** for errors and conversions

### Phase 2: Feature Completion
1. ✅ **Add remaining encoders** (Form if needed)
2. ✅ **Implement file upload support** (if prioritized)
3. ✅ **Add comprehensive stdlib testing**
4. ✅ **Create usage examples and documentation**
5. ✅ **Performance optimization**

### Phase 3: WASM Integration  
1. ✅ **WASM-specific implementation completion**
2. ✅ **Playwright-go testing setup**
3. ✅ **Cross-platform integration tests**
4. ✅ **Performance comparison and optimization**

### Phase 4: Migration and Polish
1. ✅ **Migration guide from fetchclient/fetchserver**
2. ✅ **Backward compatibility considerations**
3. ✅ **Final documentation and examples**
4. ✅ **Release preparation**

---

**Note**: This proposal is open for discussion and modification. All stakeholders should review and provide feedback on the proposed architecture and design decisions before implementation begins.
