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

	"github.com/osapi-io/osapi/internal/authtoken"
	"gopkg.in/yaml.v3"
)

// signingKey is the key of controller.api.security.signing_keys in
// osapi.yaml here.
var signingKey = authtoken.Key{
	ID:     "integration",
	Secret: "1dcf49fb9ad538f8b70fdb4ccb73c78692022e3191dba86f63308f0c3ee0e3c3",
}

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
	// controllerURL is the controller the suite calls: its own, or the one
	// that OSAPI_INTEGRATION_OSAPIFILE names, such as the lab's.
	controllerURL string
	// target is where a test that names one host sends its job:
	// OSAPI_INTEGRATION_TARGET, or _any. Against several agents, such as the
	// lab's, it must be one host, so a read and the write after it reach the
	// same agent.
	target string
	// proxyPort is the port of the spec proxy, which the CLI calls.
	proxyPort int
	proxy     *specProxy
)

func TestMain(
	m *testing.M,
) {
	var err error

	runWrites = os.Getenv("OSAPI_INTEGRATION_WRITES") == "1"
	target = os.Getenv("OSAPI_INTEGRATION_TARGET")
	if target == "" {
		target = "_any"
	}

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

	configPath, err = filepath.Abs(filepath.Join(repoRoot(), "test", "integration", "osapi.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve config path: %v\n", err)
		os.Exit(1)
	}

	if file := os.Getenv("OSAPI_INTEGRATION_OSAPIFILE"); file != "" {
		os.Exit(runAgainst(m, file))
	}

	t := authtoken.New(nil)
	token, err = t.Generate(signingKey, []string{"admin"}, "integration@test", nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate token: %v\n", err)
		os.Exit(1)
	}

	controllerURL = fmt.Sprintf("http://127.0.0.1:%d", apiPort)
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

	proxy, proxyPort, err = startSpecProxy(controllerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start the spec proxy: %v\n", err)
		stopServer()
		os.Exit(1)
	}

	// The controller is ready before the agent's first heartbeat. A suite
	// that runs first then finds no agent (#806). A failed first heartbeat
	// is written again after the 10s heartbeat interval.
	if err := waitForAgent(30 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for agent: %v\n", err)
		stopServer()
		os.Exit(1)
	}

	code := m.Run()

	// Each request and response went through the spec: one that broke it
	// fails the run, even when its own test passed.
	if v := proxy.Violations(); len(v) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d requests or responses break the full spec:\n", len(v))
		for _, line := range v {
			fmt.Fprintln(os.Stderr, "  "+line)
		}
		code = 1
	}

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
	return append(
		os.Environ(),
		fmt.Sprintf("OSAPI_NATS_SERVER_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_NATS_SERVER_STORE_DIR=%s", storeDir),
		fmt.Sprintf("OSAPI_CONTROLLER_API_PORT=%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_AGENT_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}

func clientEnv() []string {
	return append(
		os.Environ(),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_URL=http://127.0.0.1:%d", proxyPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
}

func waitForReady(
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	url := controllerURL + "/health/ready"

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

// waitForAgent waits until agent list has at least one agent.
func waitForAgent(
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		stdout, _, exitCode := runCLI("client", "agent", "list", "--json")
		if exitCode == 0 {
			var resp struct {
				Items []json.RawMessage `json:"items"`
			}
			if parseJSON(stdout, &resp) == nil && len(resp.Items) > 0 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("no agent registered after %s", timeout)
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
		t.Skip("skipping write test (set OSAPI_INTEGRATION_WRITES=1 to enable)")
	}
}

// skipWriteOp skips a mutating test unless the global OSAPI_INTEGRATION_WRITES=1
// or the operation-specific override is set (e.g., OSAPI_INTEGRATION_WRITE_HOSTNAME_UPDATE=1).
func skipWriteOp(
	t *testing.T,
	op string,
) {
	t.Helper()

	envVar := "OSAPI_INTEGRATION_WRITE_" + op
	if !runWrites && os.Getenv(envVar) != "1" {
		t.Skipf(
			"skipping write test (set OSAPI_INTEGRATION_WRITES=1 or %s=1)",
			envVar,
		)
	}
}

func writeTempFile(
	t *testing.T,
	content string,
) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}

// runAgainst runs the suite against a controller that is already running,
// such as the lab's, and returns its exit code. file is an osapi client
// config: its controller.client URL and bearer token are the controller and
// the token. Nothing starts here, and nothing stops: the controller and its
// agents are the ones the file names. The spec proxy still checks every
// request and response.
func runAgainst(
	m *testing.M,
	file string,
) int {
	var cfg struct {
		Controller struct {
			Client struct {
				URL      string `yaml:"url"`
				Security struct {
					BearerToken string `yaml:"bearer_token"`
				} `yaml:"security"`
			} `yaml:"client"`
		} `yaml:"controller"`
	}
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", file, err)
		return 1
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", file, err)
		return 1
	}
	controllerURL = strings.TrimSuffix(cfg.Controller.Client.URL, "/")
	token = cfg.Controller.Client.Security.BearerToken
	if controllerURL == "" || token == "" {
		fmt.Fprintf(os.Stderr, "%s has no controller.client url or bearer_token\n", file)
		return 1
	}
	fmt.Fprintf(os.Stderr, "integration: against %s, target %s\n", controllerURL, target)

	if err := waitForReady(15 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for ready: %v\n", err)
		return 1
	}
	proxy, proxyPort, err = startSpecProxy(controllerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start the spec proxy: %v\n", err)
		return 1
	}
	if err := waitForAgent(30 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for agent: %v\n", err)
		return 1
	}

	code := m.Run()
	if v := proxy.Violations(); len(v) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d requests or responses break the full spec:\n", len(v))
		for _, line := range v {
			fmt.Fprintln(os.Stderr, "  "+line)
		}
		code = 1
	}
	os.RemoveAll(tempDir)

	return code
}

// skipIfCommandOff skips a command test when the agent has
// agent.command.enabled off, as the lab's agents do. The suite's own agent
// has it on.
func skipIfCommandOff(
	t *testing.T,
	output string,
) {
	t.Helper()
	if strings.Contains(output, "command execution is disabled") {
		t.Skip("the agent has agent.command.enabled off")
	}
}
