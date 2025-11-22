# Implementation Guide

This guide provides a step-by-step checklist for refactoring the `fetchgo` library to the new architecture.

## Phase 1: Core Architecture

**Goal**: Set up the new `Fetchgo` manager and `Client` structure with absolute URLs.

- [ ] **Modify `fetchgo.go`**:
    - [ ] Redefine the `Fetchgo` struct to include `tb *tinybin.TinyBin`, `corsMode string`, and `corsCredentials bool`.
    - [ ] Implement the `New() *Fetchgo` constructor to initialize default values.
    - [ ] Implement the `SetCORS(mode string, credentials bool) *Fetchgo` method (chainable).
    - [ ] Implement `NewClient(timeoutMS int) Client` method (NO baseURL parameter).
    - [ ] Implement the internal `getJSONEncoder() encoder` method (platform-specific).
    - [ ] Implement the internal `getTinyBinEncoder() encoder` method (returns tinyBinEncoder with tb).

- [ ] **Modify `client.go`**:
    - [ ] Define the private `client` struct (lowercase `c`). It should contain ONLY: `defaultHeaders map[string]string`, `timeoutMS int`, and `*Fetchgo` reference.
    - [ ] **IMPORTANT**: Remove any `baseURL` field. All URLs will be absolute.
    - [ ] Ensure `client` struct implements the `Client` interface.
    - [ ] Implement `SendJSON(method, url string, body any, callback func([]byte, error))`:
        - Calls `doRequest()` with JSON encoder and `"application/json"` content-type
        - Callback receives raw `[]byte` response body
        - User decodes with `json.Unmarshal`
    - [ ] Implement `SendBinary(method, url string, body any, callback func([]byte, error))`:
        - Calls `doRequest()` with TinyBin encoder and `"application/octet-stream"` content-type
        - Callback receives raw `[]byte` response body
        - User decodes with `tinybin.Decode`
    - [ ] Implement `SetHeader(key, value string)` (void, not chainable).

- [ ] **Modify `types.go` (or create `interfaces.go`)**:
    - [ ] Define the public `Client` interface with explicit methods:
        - `SendJSON(method, url string, body any, callback func([]byte, error))` (JSON encoding)
        - `SendBinary(method, url string, body any, callback func([]byte, error))` (TinyBin encoding)
        - `SetHeader(key, value string)`
    - [ ] Ensure `encoder` interface is defined:
        ```go
        type encoder interface {
            Encode(data any) ([]byte, error)
        }
        ```
    - [ ] **Note**: No `decoder` interface needed. Users decode the []byte response themselves.

## Phase 2: Codec Implementation

**Goal**: Create the platform-specific and shared codec files.

- [ ] **Create `codecs_json_stdlib.go`**:
    - [ ] Add `//go:build !wasm` build tag.
    - [ ] Implement `getJSONEncoder()` for stdlib.
    - [ ] Define `stdlibJSONEncoder` and implement its `Encode(data any) ([]byte, error)` method using `json.Marshal`.

- [ ] **Create `codecs_json_wasm.go`**:
    - [ ] Add `//go:build wasm` build tag.
    - [ ] Implement `getJSONEncoder()` for WASM.
    - [ ] Define `wasmJSONEncoder` and implement its `Encode(data any) ([]byte, error)` method:
        - Convert Go value to JS using helper function
        - Use `JSON.stringify` to convert to string
        - Return as `[]byte`

- [ ] **Create `codecs_shared.go`**:
    - [ ] No build tags needed.
    - [ ] Define `tinyBinEncoder` struct with `tb *tinybin.TinyBin` field.
    - [ ] Implement its `Encode(data any) ([]byte, error)` method:
        - Calls `e.tb.Encode(data)`

- [ ] **Delete `encoders.go`**:
    - [ ] The old file with hardcoded encoders is now obsolete.

## Phase 3: Update Request Logic

**Goal**: Make the `doRequest` functions use the new automatic codec system with absolute URLs.

- [ ] **Modify `client_stdlib.go`**:
    - [ ] Update signature: `doRequest(method, url, contentType string, encoder encoder, body any) ([]byte, error)`.
    - [ ] **IMPORTANT**: Use `url` directly (it's already absolute). No baseURL concatenation.
    - [ ] Use the passed `encoder` to prepare the request body.
    - [ ] Set the `Content-Type` header from the parameter (this comes from `detectContentType()`).
    - [ ] Add all `defaultHeaders` from `c.defaultHeaders`.
    - [ ] Read the response body using `io.ReadAll(resp.Body)`.
    - [ ] Return: (response body as []byte, error).

- [ ] **Modify `client_wasm.go`**:
    - [ ] Update signature: `doRequest(method, url, contentType string, encoder encoder, body any) ([]byte, error)`.
    - [ ] **IMPORTANT**: Use `url` directly in the `fetch()` call (it's already absolute).
    - [ ] Use the passed `encoder` to prepare the request body.
    - [ ] Create headers object and set `Content-Type` from parameter (this comes from `detectContentType()`).
    - [ ] Add all `defaultHeaders` from `c.defaultHeaders` to the headers object.
    - [ ] Add the `mode` and `credentials` options to the `fetch` call from `c.fetchgo`.
    - [ ] In the promise handler:
        - Call `response.arrayBuffer()` to get the body as ArrayBuffer
        - Convert ArrayBuffer to `[]byte` using Uint8Array
    - [ ] Return: (response body as []byte, error).

## Phase 4: Documentation

**Goal**: Update the `README.md` to reflect the new, simpler API with absolute URLs.

- [ ] **Update `README.md`**:
    - [ ] Remove all references to manual encoder configuration.
    - [ ] Remove all references to `baseURL` in client creation.
    - [ ] Add a new "Quick Start" example showing:
        - `fetchgo.New().NewClient(timeout)` pattern with absolute URLs
        - Using `SendJSON()` for JSON requests
        - Using `SendBinary()` for TinyBin requests
        - **User decodes the []byte response** using `json.Unmarshal` or `tinybin.Decode`
    - [ ] Add example: Single client making requests to multiple domains with both JSON and TinyBin.
    - [ ] Add a section explaining:
        - **Two explicit methods**:
          - `SendJSON()` → JSON encoding (for third-party APIs, compatibility)
          - `SendBinary()` → TinyBin encoding (for your microservices, performance)
        - **Response decoding is manual**: User receives `[]byte` and decodes using `json.Unmarshal` or `tinybin.Decode`.
    - [ ] Emphasize: **Choose the right method** for your use case (JSON for compatibility, TinyBin for performance).
    - [ ] Add a section on how to configure CORS using `SetCORS()`.
    - [ ] Add a section on "Integration with CRUDP" showing:
        - How fetchgo serves as the transport layer
        - Example of decoding the []byte response
    - [ ] Create a simple migration guide for users of the old API:
        - Old: `client := &fetchgo.Client{BaseURL: "...", RequestType: fetchgo.RequestJSON}`
        - New: `client := fetchgo.New().NewClient(5000)` + use absolute URLs + Content-Type is auto-detected.