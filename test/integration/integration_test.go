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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

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
	os.Exit(runSuite(m))
}

// runSuite starts the suite's server, or reads the deployment that
// OSAPI_INTEGRATION_OSAPIFILE names, runs the tests, and returns the exit
// code. Its defers remove the temporary directory, stop the server and
// close the spec proxy, also when a step before the tests fails.
func runSuite(
	m *testing.M,
) int {
	var err error

	runWrites = os.Getenv("OSAPI_INTEGRATION_WRITES") == "1"
	target = os.Getenv("OSAPI_INTEGRATION_TARGET")
	if target == "" {
		target = "_any"
	}

	tempDir, err = os.MkdirTemp("", "osapi-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	storeDir = filepath.Join(tempDir, "jetstream")

	apiPort, err = getFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get free api port: %v\n", err)
		return 1
	}

	natsPort, err = getFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get free nats port: %v\n", err)
		return 1
	}

	binaryPath = filepath.Join(tempDir, "osapi")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = repoRoot()
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n", err)
		return 1
	}

	configPath, err = filepath.Abs(filepath.Join(repoRoot(), "test", "integration", "osapi.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve config path: %v\n", err)
		return 1
	}

	if file := os.Getenv("OSAPI_INTEGRATION_OSAPIFILE"); file != "" {
		return runAgainst(m, file)
	}

	// The suite's own server keeps its vault key in its directory, so a
	// user that is not root runs it too.
	configPath, err = writeConfig(tempDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	t := authtoken.New(nil)
	token, err = t.Generate(signingKey, []types.Role{types.RoleAdmin}, "integration@test", nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate token: %v\n", err)
		return 1
	}

	// The API serves TLS by default, with its self-signed api.crt in the
	// controller's key directory (osapi-io/osapi#1413).
	controllerURL = fmt.Sprintf("https://127.0.0.1:%d", apiPort)
	fmt.Fprintf(
		os.Stderr,
		"integration: api=%d nats=%d dir=%s client=%s\n",
		apiPort,
		natsPort,
		tempDir,
		clientName(),
	)

	serverCmd = exec.Command(binaryPath, "start", "-f", configPath)
	serverCmd.Env = serverEnv()
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	if err := serverCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start server: %v\n", err)
		return 1
	}
	defer stopServer()

	if err := waitForReady(15 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for ready: %v\n", err)
		return 1
	}

	proxy, proxyPort, err = startSpecProxy(controllerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start the spec proxy: %v\n", err)
		return 1
	}
	defer proxy.Close()

	// The controller is ready before the agent's first heartbeat. A suite
	// that runs first then finds no agent (#806). A failed first heartbeat
	// is written again after the 10s heartbeat interval.
	if err := waitForAgent(30 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for agent: %v\n", err)
		return 1
	}

	return finish(runPrepared(m))
}

// finish is code, or 1 when the SDK's report or the spec proxy found a
// problem: each request and response went through the spec, and one that
// broke it fails the run, even when its own test passed.
func finish(
	code int,
) int {
	if reportSDK() != 0 {
		code = 1
	}
	if v := proxy.Violations(); len(v) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d requests or responses break the full spec:\n", len(v))
		for _, line := range v {
			fmt.Fprintln(os.Stderr, "  "+line)
		}
		code = 1
	}

	return code
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
		fmt.Sprintf("OSAPI_NATS_SERVER_KEY_DIR=%s", filepath.Join(tempDir, "pki", "nats")),
		fmt.Sprintf("OSAPI_CONTROLLER_API_PORT=%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_AGENT_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
		// PKI is always on: the controller and the agent keep their keys in
		// the suite's directory, not in /etc/osapi/pki.
		fmt.Sprintf("OSAPI_CONTROLLER_PKI_KEY_DIR=%s", filepath.Join(tempDir, "pki", "controller")),
		fmt.Sprintf("OSAPI_AGENT_PKI_KEY_DIR=%s", filepath.Join(tempDir, "pki", "agent")),
	)
}

func clientEnv() []string {
	env := append(
		os.Environ(),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_URL=http://127.0.0.1:%d", proxyPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", token),
	)
	if os.Getenv("OSAPI_INTEGRATION_OSAPIFILE") != "" {
		return env
	}

	return append(env, hostEnv()...)
}

// hostEnv is what a command that runs on the host of the suite's server
// reads besides the client's settings, such as join token create and nats
// setup: the key directories of the server, its NATS port, and the addresses
// that a join string names. Without them such a command reads
// /etc/osapi/pki, which the runner does not have.
func hostEnv() []string {
	return []string{
		fmt.Sprintf("OSAPI_NATS_SERVER_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_NATS_SERVER_KEY_DIR=%s", filepath.Join(tempDir, "pki", "nats")),
		fmt.Sprintf("OSAPI_NATS_SERVER_URL=tls://localhost:%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_API_URL=https://127.0.0.1:%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_PKI_KEY_DIR=%s", filepath.Join(tempDir, "pki", "controller")),
		fmt.Sprintf("OSAPI_AGENT_PKI_KEY_DIR=%s", filepath.Join(tempDir, "pki", "agent")),
	}
}

// apiClient is an HTTP client that trusts the API of the suite: its
// self-signed api.crt, once the controller made it, besides the system's
// roots.
func apiClient() *http.Client {
	return apiClientFor(filepath.Join(tempDir, "pki", "controller", "api.crt"))
}

// certTransport is a transport that trusts the certificate in file as it
// is now: a rotation writes a new api.crt, which the running controller
// serves at once, so a client that the suite keeps, such as the spec
// proxy's, reads the file again when it changed (osapi-io/osapi#1569).
type certTransport struct {
	file string

	mu  sync.Mutex
	pem []byte
	t   http.RoundTripper
}

func (c *certTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	pem, _ := os.ReadFile(c.file)
	c.mu.Lock()
	if c.t == nil || !bytes.Equal(pem, c.pem) {
		c.t, c.pem = apiClientFor(c.file).Transport, pem
	}
	t := c.t
	c.mu.Unlock()

	return t.RoundTrip(req)
}

// apiClientFor is an HTTP client that trusts the certificate in certFile,
// once a controller made it, besides the system's roots.
func apiClientFor(
	certFile string,
) *http.Client {
	pool, _ := x509.SystemCertPool()
	if pool == nil {
		pool = x509.NewCertPool()
	}
	if pem, err := os.ReadFile(certFile); err == nil {
		pool.AppendCertsFromPEM(pem)
	}
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}

	return &http.Client{Transport: t, Timeout: 10 * time.Second}
}

func waitForReady(
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	url := controllerURL + "/api/health/ready"

	for time.Now().Before(deadline) {
		resp, err := apiClient().Get(url)
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

// waitForAgent waits until agent list has at least one agent that the
// controller admitted: an agent takes no job while it is Pending.
func waitForAgent(
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		stdout, _, exitCode := runCLI("client", "agent", "list", "--json")
		if exitCode == 0 {
			var resp struct {
				Items []struct {
					State string `json:"state"`
				} `json:"items"`
			}
			if parseJSON(stdout, &resp) == nil {
				for _, a := range resp.Items {
					if a.State != "Pending" {
						return nil
					}
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("no admitted agent after %s", timeout)
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
	return runCLIStdin("", args...)
}

// runCLIStdin is runCLI with stdin as the command's standard input, which a
// secret flag of - reads.
func runCLIStdin(
	stdin string,
	args ...string,
) (string, string, int) {
	return runCLIEnv(stdin, nil, args...)
}

// runCLIEnv is runCLIStdin with env added to the client's environment,
// such as the NATS user that a command connects as.
func runCLIEnv(
	stdin string,
	env []string,
	args ...string,
) (string, string, int) {
	fullArgs := append([]string{"-f", configPath}, args...)
	cmd := exec.Command(binaryPath, fullArgs...)
	cmd.Env = append(clientEnv(), env...)
	if hostTarget != nil {
		cmd = exec.Command(binaryPath, append([]string{"-f", hostTarget.cfg}, args...)...)
		cmd.Env = append(append([]string{}, hostTarget.env...), env...)
	}
	cmd.Stdin = strings.NewReader(stdin)

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

// runPrepared runs the tests on a target that prepare made ready, when the
// write examples run, and puts back what prepare changed after them. A
// prepare that fails stops the run: each test would fail for what is
// missing, not for what it tests.
func runPrepared(
	m *testing.M,
) int {
	if !prepareWrites() {
		return m.Run()
	}
	if err := prepare(); err != nil {
		fmt.Fprintf(os.Stderr, "prepare the target: %v\n", err)
		_ = unprepare()

		return 1
	}
	code := m.Run()
	if err := unprepare(); err != nil {
		fmt.Fprintf(os.Stderr, "put back the target: %v\n", err)
		code = 1
	}

	return code
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
	fmt.Fprintf(
		os.Stderr,
		"integration: against %s, target %s, client %s\n",
		controllerURL,
		target,
		clientName(),
	)

	if err := waitForReady(15 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for ready: %v\n", err)
		return 1
	}
	proxy, proxyPort, err = startSpecProxy(controllerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start the spec proxy: %v\n", err)
		return 1
	}
	defer proxy.Close()
	if err := waitForAgent(30 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for agent: %v\n", err)
		return 1
	}

	return finish(runPrepared(m))
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
