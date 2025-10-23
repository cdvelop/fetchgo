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
**Decision**: **Pluggable encoder Interface**

**Implementation**:
```go
type encoder interface {
    Encode(any) ([]byte, error)
    Decode([]byte, &any) error
}

type Client struct {
    encoder encoder // Optional, defaults NONE
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
    encoder        encoder
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

### 1. 🔧 **TinyGo Header Handling Strategy** ✅ DECIDED
**Decision**: Minimal helper methods (`AddHeader`, `SetHeader`) will be provided for convenience. Direct slice manipulation remains possible for advanced use cases. The internal `upsertHeader` and `getHeaders` logic is approved as proposed.

**Final Implementation**:
```go
// Public helpers approved
func (c *Client) AddHeader(key, value string) { c.upsertHeader(key, value, false) }
func (c *Client) SetHeader(key, value string) { c.upsertHeader(key, value, true) }

// Internal logic approved
// func (c *Client) upsertHeader(...)
// func (c *Client) getHeaders() map[string]string
```
**Consideration**: No validation will be performed on the raw `DefaultHeaders` slice. It is the developer's responsibility to ensure it contains an even number of elements.

### 2. ⏱️ **Environment-Specific Timeout Handling** ✅ DECIDED
**Decision**: **Option A** (Build-tag specific implementation) is approved. This approach is simple, avoids over-engineering, and cleanly separates the environment-specific logic.

**Final Implementation**:
```go
// client_wasm.go
// func (c *Client) applyTimeout(timeoutMS int) js.Value { ... }

// client_stdlib.go
// func (c *Client) applyTimeout(timeoutMS int) context.Context { ... }
```

### 3. 📝 **Request Type Handling** ✅ DECIDED
**Decision**: **Option A** (Per-client `RequestType`) is approved for the initial version. This keeps the API simple. The ability to override on a per-request basis can be added later if needed.

**Final Implementation**: The `requestType` will be a field on the `Client` struct and will determine the `Content-Type` header and body encoding for all requests made with that client instance.

### 4. 🔌 **encoder Interface Implementation** ✅ DECIDED
**Decision**: **Option C** (JSON and Raw encoders) is approved. This provides a balanced set of default encoders that cover the most common scenarios (structured data and binary/pass-through data) without adding unnecessary complexity.

**Final Implementation**:
- `JSONEncoder`: Default for `RequestJSON`.
- `RawEncoder`: Default for `RequestRaw`.
- The `encoder` interface and pluggable design are approved.

### 5. � File and Binary Payload Handling ✅ DECIDED

**Decision**: The proposal to consolidate file and binary handling into the main `SendRequest` function is approved. This maintains a minimal API surface.

**Final Implementation**: `SendRequest` will accept `[]byte`, `string` (as a file path for StdLib), and `io.Reader` for StdLib, and `js.Value` (e.g., File/Blob objects) for WASM, as outlined in the proposal. The internal `prepareBody` helper is a good approach for implementation.

### 6. 🎯 **SendRequest Method Versatility** ✅ DECIDED
**Decision**: The versatile signature for `SendRequest` is approved. The internal processing logic is a solid plan. To improve robustness, basic validation will be added to ensure the `method` parameter is a standard HTTP method.

**Final Implementation**:
- The proposed `SendRequest` signature and internal logic are approved.
- An internal check will be added to validate the `method` string (e.g., GET, POST, etc.) and return an error for invalid methods.
- Special handling for HEAD requests (discarding the response body) will be implemented.
- Auto-retry logic is considered out of scope for the initial version.

### 7. 📦 **Dependency Management** ✅ DECIDED
**Decision**: The integration of `github.com/cdvelop/tinystring` is approved. It will be used for all internal error creation and string conversions to ensure the library is lightweight and optimized for TinyGo.

**Final Implementation**: The library will import `github.com/cdvelop/tinystring` and use its `Err`, `Errf`, and `Convert` functions. No wrappers are necessary.

### 8. 🗂️ **File Structure Refinement** ✅ DECIDED
**Decision**: The proposed file structure is approved with a modification: timeout and file upload logic will be integrated directly into the `client_wasm.go` and `client_stdlib.go` files instead of separate `timeout_*.go` or `upload.go` files. This reduces the number of files and keeps related logic co-located.

**Final Structure**:
```
fetchgo/
├── client.go          // Main Client struct and public API (SendRequest)
├── types.go           // Public types (requestType) and interfaces (encoder)
├── encoders.go        // Default encoder implementations (JSON, Raw)
├── client_wasm.go     // WASM-specific implementation (doRequest, timeout)
├── client_stdlib.go   // StdLib implementation (doRequest, timeout, file handling)
├── headers.go         // Shared utilities for header processing
└── url.go             // Shared utilities for URL building
```

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
1. **DataConverter.EncodeMaps()** → Pluggable encoder interface
2. **DataConverter.DecodeMaps()** → Pluggable encoder interface  
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
3. ✅ **Create encoder interface** and JSON/Raw implementations
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
