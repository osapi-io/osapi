//go:build integration

// Copyright (c) 2024 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/retr0h/osapi/internal/authtoken"
)

const signingKey = "111fdb0cfd9788fa6af8815f856a0374bf7a0174ad62fa8b98ec07a55f68d8d8"

var (
	binaryPath string
	apiPort    int
	natsPort   int
	token      string
	configPath string
	serverCmd  *exec.Cmd
	tempDir    string
	storeDir   string
	runWrites  bool
)

func TestMain(
	m *testing.M,
) {
	var err error

	runWrites = os.Getenv("INTEGRATION_WRITES") == "1"

	tempDir, err = os.MkdirTemp("", "osapi-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}

	storeDir = filepath.Join(tempDir, "jetstream")

	apiPort, err = getFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get free api port: %v\n", err)
		os.Exit(1)
	}

	natsPort, err = getFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get free nats port: %v\n", err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tempDir, "osapi")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = repoRoot()
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n", err)
		os.Exit(1)
	}

	t := authtoken.New(nil)
	token, err = t.Generate(signingKey, []string{"admin"}, "integration@test", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate token: %v\n", err)
		os.Exit(1)
	}

	configPath, err = filepath.Abs(filepath.Join(repoRoot(), "test", "integration", "osapi.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve config path: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "integration: api=%d nats=%d dir=%s\n", apiPort, natsPort, tempDir)

	serverCmd = exec.Command(binaryPath, "start", "-f", configPath)
	serverCmd.Env = serverEnv()
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start server: %v\n", err)
		os.Exit(1)
	}

	if err := waitForReady(15 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for ready: %v\n", err)
		stopServer()
		os.Exit(1)
	}

	code := m.Run()

	stopServer()
	os.RemoveAll(tempDir)
	os.Exit(code)
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("listen for free port: %w", err)
	}
	defer l.Close()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected addr type: %T", l.Addr())
	}

	return addr.Port, nil
}

func repoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get working directory: %v", err))
	}

	// test/integration/ is two levels below repo root
	return filepath.Join(wd, "..", "..")
}

func serverEnv() []string {
	return append(os.Environ(),
		fmt.Sprintf("OSAPI_NATS_SERVER_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_NATS_SERVER_STORE_DIR=%s", storeDir),
		fmt.Sprintf("OSAPI_API_SERVER_PORT=%d", apiPort),
		fmt.Sprintf("OSAPI_API_SERVER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_AGENT_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}

func clientEnv() []string {
	return append(os.Environ(),
		fmt.Sprintf("OSAPI_API_CLIENT_URL=http://127.0.0.1:%d", apiPort),
		fmt.Sprintf("OSAPI_API_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}

func waitForReady(
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/health/ready", apiPort)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:gosec
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("server not ready after %s", timeout)
}

func stopServer() {
	if serverCmd != nil && serverCmd.Process != nil {
		_ = serverCmd.Process.Kill()
		_ = serverCmd.Wait()
	}
}

func runCLI(
	args ...string,
) (string, string, int) {
	fullArgs := append([]string{"-f", configPath}, args...)
	cmd := exec.Command(binaryPath, fullArgs...)
	cmd.Env = clientEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

func parseJSON(
	raw string,
	target any,
) error {
	return json.Unmarshal([]byte(strings.TrimSpace(raw)), target)
}

func skipWrite(
	t *testing.T,
) {
	t.Helper()
	if !runWrites {
		t.Skip("skipping write test (set INTEGRATION_WRITES=1 to enable)")
	}
}
