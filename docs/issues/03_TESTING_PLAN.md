# Testing Plan

**Goal**: Achieve >90% test coverage by implementing a comprehensive test suite that validates all aspects of the new architecture.

## Phase 1: Basic Setup & Unit Tests

- [ ] **Test `fetchgo.go`**:
    - [ ] Test `New()`: Verify that a new `Fetchgo` instance has the correct default values (`corsMode: "cors"`, `corsCredentials: false`, `tb` is not nil).
    - [ ] Test `SetCORS()`:
        - Verify that `SetCORS("cors", true)` correctly updates the `corsMode` and `corsCredentials` fields.
        - Verify that the method is chainable.
    - [ ] Test `NewClient()`:
        - Verify that it returns a non-nil `Client` interface.
        - Verify the underlying `*client` struct has the correct `timeout` and a reference to the parent `Fetchgo` instance.
        - **IMPORTANT**: Verify there is NO `baseURL` field in the struct.
    - [ ] Test `getJSONEncoder()`:
        - Verify it returns a valid JSON encoder.
        - Test encoding a struct returns valid JSON bytes.
    - [ ] Test `getTinyBinEncoder()`:
        - Verify it returns a `tinyBinEncoder` with valid `tb` instance.
        - Test encoding a struct returns valid TinyBin bytes.
    - [ ] **Note**: No `getDecoder()` tests needed. Responses are returned as raw `[]byte`.

- [ ] **Test Codecs**:
    - [ ] Test `stdlibJSONEncoder` (`!wasm`): Verify that a Go struct can be encoded to JSON `[]byte`.
    - [ ] Test `wasmJSONEncoder` (`wasm`): Verify that a Go struct can be encoded and returns valid JSON as `[]byte`.
    - [ ] Test `tinyBinEncoder`: Verify that a Go struct can be encoded to TinyBin `[]byte`.

## Phase 2: End-to-End (E2E) Request Tests

These tests should run against the live test server (`testserver/main.go`) in both `stdlib` and `wasm` environments using the shared test file structure.

- [ ] **Update `server_test.go` / `testserver/main.go`**:
    - [ ] Add a new endpoint `/echo` that reads the request body and writes it back to the response, preserving the `Content-Type` header.
    - [ ] **IMPORTANT**: You will need to run TWO test servers for multi-domain testing:
        - Server 1: `localhost:8080` (main test server)
        - Server 2: `localhost:9090` (secondary for multi-domain tests)

- [ ] **Create `fetch_shared_test.go` tests**:
    - [ ] **Test SendJSON E2E**:
        - Use `SendJSON()` to send a struct to `http://localhost:8080/echo`.
        - Verify that the response callback receives `[]byte`.
        - Use `json.Unmarshal` to decode the `[]byte` back to the original struct.
        - Verify the server received `Content-Type: application/json` header.
    - [ ] **Test SendBinary E2E**:
        - Use `SendBinary()` to send a struct to `http://localhost:8080/echo`.
        - Verify that the response callback receives `[]byte`.
        - Use `tinybin.Decode` to decode the `[]byte` back to the original struct.
        - Verify the server received `Content-Type: application/octet-stream` header.
    - [ ] **Test GET Request (No Body)**:
        - Send a GET request with `body: nil` to `http://localhost:8080/health`.
        - Verify a `200 OK` status and no Content-Type header sent.
    - [ ] **Test Custom Headers**:
        - Use `SetHeader()` to add a custom header (e.g., `X-Custom-Header: my-value`).
        - Send request to `http://localhost:8080/echo` and verify the response includes the header.
    - [ ] **Test Multiple Domains (Critical for Multi-Domain Support)**:
        - Create a single client.
        - Make a `SendJSON()` request to `http://localhost:8080/echo` with a struct.
        - In the callback, decode the `[]byte` response using `json.Unmarshal`.
        - Make a `SendBinary()` request to `http://localhost:9090/echo` with a struct.
        - In the callback, decode the `[]byte` response using `tinybin.Decode`.
        - Verify both callbacks receive `[]byte`.
        - Verify both succeed without creating multiple clients.
        - Verify correct Content-Types were sent for each request.
    - [ ] **Test Explicit Method Selection**:
        - Test `SendJSON()` always sends `Content-Type: application/json`.
        - Test `SendBinary()` always sends `Content-Type: application/octet-stream`.
        - Test both methods work with structs, maps, and slices.
    - [ ] **Test Error Handling**:
        - Test a request to `http://localhost:8080/not-found` and verify 404 error.
        - Test a request to an invalid domain `http://invalid.local` and verify network error.
    - [ ] **Test Timeout**:
        - Set a very short timeout (e.g., 1ms) for a request to a slow endpoint.
        - Verify that the error callback is triggered with a timeout error.
    - [ ] **Test CORS (WASM only)**:
        - Use `SetCORS("cors", true)`.
        - Send a request from the WASM test environment.
        - Verify that the request succeeds with the correct CORS configuration.

## Phase 3: Cleanup

- [ ] **Review and Refactor**:
    - [ ] Ensure all public functions and types in the library are covered by tests.
    - [ ] Remove any old, now-unused test files or helper functions.
    - [ ] Run `go test -cover` to confirm that coverage is above 90%.