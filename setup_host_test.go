//go:build !wasm

package fetch_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// 1. Check if server is running
	if !isServerRunning() {
		fmt.Println("Starting test server...")
		if err := startServer(); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Test server already running.")
	}

	// 2. Run tests
	code := m.Run()

	// 3. Exit (Server remains running for WASM tests)
	os.Remove("testserver/testserver_bin")
	os.Exit(code)
}

func isServerRunning() bool {
	urlBytes, err := os.ReadFile(".test_server_url")
	if err != nil {
		return false
	}
	serverURL := string(urlBytes)
	resp, err := http.Get(serverURL + "/get")
	if err != nil {
		// Server URL file exists but server is not responding, clean up stale file
		os.Remove(".test_server_url")
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func startServer() error {
	// Remove old URL file
	os.Remove(".test_server_url")

	// Build server first
	buildCmd := exec.Command("go", "build", "-o", "testserver/testserver_bin", "./testserver")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build test server: %v, output: %s", err, string(out))
	}

	// Start server detached - use Setsid to create new session so it survives parent exit
	cmd := exec.Command("./testserver_bin")
	cmd.Dir = "testserver"
	// Redirect to /dev/null to prevent SIGPIPE when parent exits
	devNull, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0)
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	// Create new session so server survives test process exit
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for URL file
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(".test_server_url"); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for server to start")
}
