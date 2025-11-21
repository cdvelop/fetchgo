#!/bin/bash

echo "=========================================="
echo "Running Stdlib Tests..."
echo "=========================================="
go test -v ./...

if [ $? -ne 0 ]; then
    echo "❌ Stdlib tests failed"
    exit 1
fi

echo ""
echo "=========================================="
echo "Starting test server for WASM tests..."
echo "=========================================="

# Start test server in background
(cd testserver && go run main.go) &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo ""
echo "=========================================="
echo "Running WASM Tests..."
echo "=========================================="

# Check if wasmbrowsertest is installed
if ! command -v wasmbrowsertest &> /dev/null; then
    echo "⚠️  wasmbrowsertest not found. Install it with:"
    echo "   go install github.com/agnivade/wasmbrowsertest@latest"
    echo "   export PATH=\$PATH:\$(go env GOPATH)/bin"
    kill $SERVER_PID
    exit 1
fi

# Run WASM tests (excluding testserver directory)
GOOS=js GOARCH=wasm go test -v -tags wasm 2>&1 | grep -v "ERROR: could not unmarshal"
WASM_EXIT_CODE=$?

# Stop test server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

if [ $WASM_EXIT_CODE -ne 0 ]; then
    echo ""
    echo "❌ WASM tests failed"
    exit 1
fi

echo ""
echo "✅ All tests passed!"