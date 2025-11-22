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

# Remove old URL file if exists
rm -f .test_server_url

# Start test server in background
(cd testserver && go run main.go) &
SERVER_PID=$!

# Wait for server to start and write URL file
for i in {1..30}; do
    if [ -f .test_server_url ]; then
        break
    fi
    sleep 0.1
done

if [ ! -f .test_server_url ]; then
    echo "❌ Server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

SERVER_URL=$(cat .test_server_url)
echo "Server started at: $SERVER_URL"

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

# Clean up URL file
rm -f .test_server_url

if [ $WASM_EXIT_CODE -ne 0 ]; then
    echo ""
    echo "❌ WASM tests failed"
    exit 1
fi

echo ""
echo "✅ All tests passed!"