# Agent Privilege Escalation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add config-driven sudo escalation for write commands, Linux capability
verification, and startup preflight checks to the OSAPI agent so it can run as
an unprivileged user.

**Architecture:** The exec `Manager` interface gains `RunPrivilegedCmd` which
prepends `sudo` when configured. Providers call it for write operations. At
startup, the agent runs preflight checks to verify sudo and capabilities before
accepting jobs.

**Tech Stack:** Go, Linux capabilities (`/proc/self/status`), sudo

---

### Task 1: Add privilege escalation config

**Files:**

- Modify: `internal/config/types.go:353-374`
- Modify: `configs/osapi.yaml`
- Modify: `configs/osapi.nerd.yaml`
- Modify: `configs/osapi.local.yaml`

- [ ] **Step 1: Add PrivilegeEscalation struct and wire into AgentConfig**

In `internal/config/types.go`, add a new struct before `AgentConfig`:

```go
// PrivilegeEscalation configuration for least-privilege agent mode.
type PrivilegeEscalation struct {
	// Sudo prepends "sudo" to write commands when true.
	Sudo bool `mapstructure:"sudo"`
	// Capabilities verifies Linux capabilities at startup when true.
	Capabilities bool `mapstructure:"capabilities"`
	// Preflight runs privilege checks before accepting jobs when true.
	Preflight bool `mapstructure:"preflight"`
}
```

Add the field to `AgentConfig`:

```go
type AgentConfig struct {
	// ... existing fields ...
	// PrivilegeEscalation configures least-privilege agent mode.
	PrivilegeEscalation PrivilegeEscalation `mapstructure:"privilege_escalation,omitempty"`
}
```

- [ ] **Step 2: Add config to YAML files**

In `configs/osapi.yaml`, `configs/osapi.nerd.yaml`, and
`configs/osapi.local.yaml`, add to the `agent:` section (disabled by default):

```yaml
# Least-privilege mode. When enabled, the agent runs as an
# unprivileged user and uses sudo for write operations.
# privilege_escalation:
#   sudo: false
#   capabilities: false
#   preflight: false
```

- [ ] **Step 3: Verify build**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```
feat(agent): add privilege escalation config
```

---

### Task 2: Add RunPrivilegedCmd to exec manager

**Files:**

- Modify: `internal/exec/manager.go`
- Modify: `internal/exec/types.go`
- Modify: `internal/exec/exec.go`
- Create: `internal/exec/run_privileged_cmd.go`
- Create: `internal/exec/run_privileged_cmd_public_test.go`
- Modify: `internal/exec/mocks/generate.go`

- [ ] **Step 1: Add RunPrivilegedCmd to the Manager interface**

In `internal/exec/manager.go`:

```go
type Manager interface {
	// RunCmd executes the provided command with arguments, using the
	// current working directory. Use for read operations.
	RunCmd(
		name string,
		args []string,
	) (string, error)

	// RunPrivilegedCmd executes a command with privilege escalation.
	// When sudo is enabled in config, prepends "sudo" to the command.
	// When sudo is disabled, behaves identically to RunCmd.
	// Use for write operations that modify system state.
	RunPrivilegedCmd(
		name string,
		args []string,
	) (string, error)

	// RunCmdFull executes a command with separate stdout/stderr capture,
	// an optional working directory, and a timeout in seconds.
	RunCmdFull(
		name string,
		args []string,
		cwd string,
		timeout int,
	) (*CmdResult, error)
}
```

- [ ] **Step 2: Add sudo field to Exec struct and update constructor**

In `internal/exec/types.go`, add the `sudo` field:

```go
type Exec struct {
	logger *slog.Logger
	sudo   bool
}
```

In `internal/exec/exec.go`, update the constructor:

```go
func New(
	logger *slog.Logger,
	sudo bool,
) *Exec {
	return &Exec{
		logger: logger.With(slog.String("subsystem", "exec")),
		sudo:   sudo,
	}
}
```

- [ ] **Step 3: Create RunPrivilegedCmd implementation**

Create `internal/exec/run_privileged_cmd.go`:

```go
package exec

// RunPrivilegedCmd executes a command with privilege escalation.
// When sudo is enabled, the command is run via "sudo". When disabled,
// it behaves identically to RunCmd.
func (e *Exec) RunPrivilegedCmd(
	name string,
	args []string,
) (string, error) {
	if e.sudo {
		args = append([]string{name}, args...)
		name = "sudo"
	}

	return e.RunCmdImpl(name, args, "")
}
```

- [ ] **Step 4: Write tests**

Create `internal/exec/run_privileged_cmd_public_test.go`. Test:

- When sudo is false, RunPrivilegedCmd behaves like RunCmd (runs the command
  directly)
- When sudo is true, the command is prefixed with sudo (verify the actual args
  passed to the underlying exec)

Since `RunCmdImpl` calls `os/exec.Command`, use a test helper pattern: override
the command execution to capture what would be run. Look at how
`run_cmd_public_test.go` tests `RunCmd` and follow the same pattern.

- [ ] **Step 5: Update agent_setup.go to pass sudo config**

In `cmd/agent_setup.go`, change the `exec.New` call:

```go
// Before:
execManager := exec.New(log)

// After:
execManager := exec.New(
	log,
	appConfig.Agent.PrivilegeEscalation.Sudo,
)
```

- [ ] **Step 6: Regenerate mocks**

Run: `go generate ./internal/exec/mocks/...`

- [ ] **Step 7: Fix compilation errors**

The mock `Manager` now requires `RunPrivilegedCmd`. Any existing test that
constructs a mock `Manager` may need updating. Run `go build ./...` and fix any
compile errors.

- [ ] **Step 8: Run tests**

Run: `go test ./internal/exec/... -count=1` Run: `go build ./...`

- [ ] **Step 9: Commit**

```
feat(exec): add RunPrivilegedCmd with config-driven sudo
```

---

### Task 3: Add preflight checks

**Files:**

- Create: `internal/agent/preflight.go`
- Create: `internal/agent/preflight_public_test.go`
- Modify: `internal/agent/server.go`

- [ ] **Step 1: Create preflight types and runner**

Create `internal/agent/preflight.go`:

```go
package agent

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
)

// PreflightResult holds the outcome of a single preflight check.
type PreflightResult struct {
	Name   string
	Passed bool
	Error  string
}

// sudoCommands lists the binaries that require sudo access.
var sudoCommands = []string{
	"systemctl",
	"sysctl",
	"timedatectl",
	"hostnamectl",
	"chronyc",
	"useradd",
	"usermod",
	"userdel",
	"groupadd",
	"groupdel",
	"gpasswd",
	"chown",
	"apt-get",
	"shutdown",
	"update-ca-certificates",
	"sh",
}

// requiredCapabilities maps capability names to their bit positions
// in the CapEff bitmask from /proc/self/status.
var requiredCapabilities = map[string]uint{
	"CAP_DAC_OVERRIDE":    1,
	"CAP_DAC_READ_SEARCH": 2,
	"CAP_FOWNER":          3,
	"CAP_KILL":            5,
}

// procStatusPath is the path to read for capability detection.
// Overridden in tests.
var procStatusPath = "/proc/self/status"

// RunPreflight checks sudo access and capabilities. Returns all
// results and whether all checks passed.
func RunPreflight(
	logger *slog.Logger,
	execManager exec.Manager,
	checkSudo bool,
	checkCaps bool,
) ([]PreflightResult, bool) {
	var results []PreflightResult
	allPassed := true

	if checkSudo {
		sudoResults := checkSudoAccess(logger, execManager)
		results = append(results, sudoResults...)
		for _, r := range sudoResults {
			if !r.Passed {
				allPassed = false
			}
		}
	}

	if checkCaps {
		capResults := checkCapabilities(logger)
		results = append(results, capResults...)
		for _, r := range capResults {
			if !r.Passed {
				allPassed = false
			}
		}
	}

	return results, allPassed
}

// checkSudoAccess verifies that sudo -n works for each required
// command. Uses "sudo -n which <command>" which is a no-op that
// tests sudo access without side effects.
func checkSudoAccess(
	logger *slog.Logger,
	execManager exec.Manager,
) []PreflightResult {
	results := make([]PreflightResult, 0, len(sudoCommands))

	for _, cmd := range sudoCommands {
		_, err := execManager.RunCmd(
			"sudo",
			[]string{"-n", "which", cmd},
		)

		result := PreflightResult{
			Name:   "sudo:" + cmd,
			Passed: err == nil,
		}
		if err != nil {
			result.Error = fmt.Sprintf(
				"sudo access denied for %s: %s",
				cmd,
				err.Error(),
			)
		}

		logger.Debug(
			"preflight sudo check",
			slog.String("command", cmd),
			slog.Bool("passed", result.Passed),
		)

		results = append(results, result)
	}

	return results
}

// checkCapabilities reads /proc/self/status and verifies that
// required capability bits are set in the effective capability mask.
func checkCapabilities(
	logger *slog.Logger,
) []PreflightResult {
	capEff, err := readCapEff()
	if err != nil {
		return []PreflightResult{{
			Name:   "capabilities",
			Passed: false,
			Error:  fmt.Sprintf("failed to read capabilities: %s", err),
		}}
	}

	results := make([]PreflightResult, 0, len(requiredCapabilities))

	for name, bit := range requiredCapabilities {
		hasCap := (capEff>>bit)&1 == 1
		result := PreflightResult{
			Name:   "cap:" + name,
			Passed: hasCap,
		}
		if !hasCap {
			result.Error = fmt.Sprintf("%s not set", name)
		}

		logger.Debug(
			"preflight capability check",
			slog.String("capability", name),
			slog.Bool("passed", hasCap),
		)

		results = append(results, result)
	}

	return results
}

// readCapEff reads the effective capability bitmask from
// /proc/self/status.
func readCapEff() (uint64, error) {
	f, err := os.Open(procStatusPath)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", procStatusPath, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CapEff:") {
			hexStr := strings.TrimSpace(
				strings.TrimPrefix(line, "CapEff:"),
			)
			bytes, err := hex.DecodeString(hexStr)
			if err != nil {
				return 0, fmt.Errorf(
					"decode CapEff %q: %w",
					hexStr,
					err,
				)
			}
			// Convert big-endian bytes to uint64.
			var val uint64
			for _, b := range bytes {
				val = (val << 8) | uint64(b)
			}
			return val, nil
		}
	}

	return 0, fmt.Errorf("CapEff not found in %s", procStatusPath)
}
```

- [ ] **Step 2: Write preflight tests**

Create `internal/agent/preflight_public_test.go`. Use testify/suite with
table-driven tests. Test:

**TestCheckSudoAccess:**

- All commands pass (mock RunCmd returns nil for all sudo -n which calls)
- One command fails (mock RunCmd returns error for one)
- Multiple commands fail

**TestCheckCapabilities:**

- All capabilities present (write a fake /proc/self/status file with full
  CapEff, override `procStatusPath` via export_test.go)
- Missing capability (write CapEff without a required bit)
- Cannot read file (set procStatusPath to nonexistent path)

**TestRunPreflight:**

- Both sudo and caps enabled and pass → allPassed true
- Sudo fails → allPassed false
- Caps fails → allPassed false
- Both disabled → empty results, allPassed true

Create `internal/agent/export_test.go` (or add to existing) to expose
`procStatusPath` for testing:

```go
package agent

func SetProcStatusPath(p string) { procStatusPath = p }
func ResetProcStatusPath()       { procStatusPath = "/proc/self/status" }
```

- [ ] **Step 3: Wire preflight into agent Start()**

In `internal/agent/server.go`, add preflight check after hostname determination
but before starting heartbeat:

```go
func (a *Agent) Start() {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.startedAt = time.Now()
	a.state = job.AgentStateReady

	a.logger.Info("starting node agent")

	a.hostname, _ = job.GetAgentHostname(a.appConfig.Agent.Hostname)

	// Run preflight checks if configured.
	pe := a.appConfig.Agent.PrivilegeEscalation
	if pe.Preflight {
		results, ok := RunPreflight(
			a.logger,
			a.execManager,
			pe.Sudo,
			pe.Capabilities,
		)
		if !ok {
			for _, r := range results {
				if !r.Passed {
					a.logger.Error(
						"preflight check failed",
						slog.String("check", r.Name),
						slog.String("error", r.Error),
					)
				}
			}
			a.logger.Error("preflight failed, agent cannot start")
			a.cancel()
			return
		}
		a.logger.Info("preflight checks passed")
	}

	// ... rest of Start() unchanged ...
```

The agent needs access to the exec manager. Check if it's already on the `Agent`
struct. If not, add an `execManager exec.Manager` field and pass it from
`agent_setup.go`.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/agent/... -count=1` Run: `go build ./...`

- [ ] **Step 5: Commit**

```
feat(agent): add preflight checks for sudo and capabilities
```

---

### Task 4: Migrate service provider to RunPrivilegedCmd

**Files:**

- Modify: `internal/provider/node/service/debian_action.go`
- Modify: `internal/provider/node/service/debian_unit.go`
- Modify: `internal/provider/node/service/debian_action_public_test.go`
- Modify: `internal/provider/node/service/debian_unit_public_test.go`

- [ ] **Step 1: Update write calls in debian_action.go**

Change these `RunCmd` calls to `RunPrivilegedCmd`:

```go
// Start — line 49: "systemctl start" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"start", unitName})

// Stop — "systemctl stop" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"stop", unitName})

// Restart — "systemctl restart" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"restart", unitName})

// Enable — "systemctl enable" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"enable", unitName})

// Disable — "systemctl disable" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"disable", unitName})
```

Keep these as `RunCmd` (reads):

- `systemctl is-active` (Start, Stop)
- `systemctl is-enabled` (Enable, Disable)

- [ ] **Step 2: Update write calls in debian_unit.go**

```go
// Delete — "systemctl stop" and "systemctl disable" are writes
d.execManager.RunPrivilegedCmd("systemctl", []string{"stop", unitName})
d.execManager.RunPrivilegedCmd("systemctl", []string{"disable", unitName})

// daemonReload — "systemctl daemon-reload" is a write
d.execManager.RunPrivilegedCmd("systemctl", []string{"daemon-reload"})
```

- [ ] **Step 3: Update test mock expectations**

In `debian_action_public_test.go`, change all mock expectations for write
commands from `RunCmd` to `RunPrivilegedCmd`. Keep read command expectations on
`RunCmd`.

In `debian_unit_public_test.go`, change mock expectations for `systemctl stop`,
`systemctl disable`, `systemctl daemon-reload` from `RunCmd` to
`RunPrivilegedCmd`.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/provider/node/service/... -count=1`

- [ ] **Step 5: Commit**

```
refactor(service): use RunPrivilegedCmd for write operations
```

---

### Task 5: Migrate sysctl provider

**Files:**

- Modify: `internal/provider/node/sysctl/debian.go`
- Modify: `internal/provider/node/sysctl/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `sysctl -p <confPath>` (apply config)
- `sysctl --system` (reload all)

Keep as `RunCmd`:

- `sysctl -n <key>` (read parameter)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/sysctl/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(sysctl): use RunPrivilegedCmd for write operations
```

---

### Task 6: Migrate hostname provider

**Files:**

- Modify: `internal/provider/node/host/debian_update_hostname.go`
- Modify: `internal/provider/node/host/debian_update_hostname_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `hostnamectl set-hostname <name>`

Keep as `RunCmd` (in other host files):

- `hostnamectl hostname` (read)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/host/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(host): use RunPrivilegedCmd for write operations
```

---

### Task 7: Migrate timezone provider

**Files:**

- Modify: `internal/provider/node/timezone/debian.go`
- Modify: `internal/provider/node/timezone/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `timedatectl set-timezone <timezone>`

Keep as `RunCmd`:

- `timedatectl show -p Timezone --value` (read)
- `date +%:z` (read)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/timezone/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(timezone): use RunPrivilegedCmd for write operations
```

---

### Task 8: Migrate NTP provider

**Files:**

- Modify: `internal/provider/node/ntp/debian.go`
- Modify: `internal/provider/node/ntp/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `chronyc reload sources`

Keep as `RunCmd`:

- `chronyc tracking` (read)
- `chronyc sources -c` (read)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/ntp/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(ntp): use RunPrivilegedCmd for write operations
```

---

### Task 9: Migrate user provider

**Files:**

- Modify: `internal/provider/node/user/debian_user.go`
- Modify: `internal/provider/node/user/debian_group.go`
- Modify: `internal/provider/node/user/debian_ssh_key.go`
- Modify: `internal/provider/node/user/debian_user_public_test.go`
- Modify: `internal/provider/node/user/debian_group_public_test.go`
- Modify: `internal/provider/node/user/debian_ssh_key_public_test.go`

- [ ] **Step 1: Update write calls in debian_user.go**

Change to `RunPrivilegedCmd`:

- `useradd --create-home ...`
- `usermod ...` (all variants)
- `userdel -r ...`
- `sh -c "echo ... | chpasswd"`

Keep as `RunCmd`:

- `id -Gn <username>` (read)
- `passwd -S <username>` (read)

- [ ] **Step 2: Update write calls in debian_group.go**

Change to `RunPrivilegedCmd`:

- `groupadd ...`
- `groupdel ...`
- `gpasswd -M ...`

- [ ] **Step 3: Update write calls in debian_ssh_key.go**

Change to `RunPrivilegedCmd`:

- `chown -R ...`

- [ ] **Step 4: Update all test mock expectations**

- [ ] **Step 5: Run tests**

Run: `go test ./internal/provider/node/user/... -count=1`

- [ ] **Step 6: Commit**

```
refactor(user): use RunPrivilegedCmd for write operations
```

---

### Task 10: Migrate package provider

**Files:**

- Modify: `internal/provider/node/apt/debian.go`
- Modify: `internal/provider/node/apt/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `apt-get install -y ...`
- `apt-get remove -y ...`
- `apt-get update`

Keep as `RunCmd`:

- `dpkg-query ...` (read)
- `apt list --upgradable` (read)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/apt/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(apt): use RunPrivilegedCmd for write operations
```

---

### Task 11: Migrate power provider

**Files:**

- Modify: `internal/provider/node/power/debian.go`
- Modify: `internal/provider/node/power/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `shutdown -r ...` (reboot)
- `shutdown -h ...` (halt)

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/power/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(power): use RunPrivilegedCmd for write operations
```

---

### Task 12: Migrate certificate provider

**Files:**

- Modify: `internal/provider/node/certificate/debian.go`
- Modify: `internal/provider/node/certificate/debian_public_test.go`

- [ ] **Step 1: Update write calls**

Change to `RunPrivilegedCmd`:

- `update-ca-certificates`

- [ ] **Step 2: Update test mock expectations**

- [ ] **Step 3: Run tests**

Run: `go test ./internal/provider/node/certificate/... -count=1`

- [ ] **Step 4: Commit**

```
refactor(certificate): use RunPrivilegedCmd for write operations
```

---

### Task 13: Migrate DNS provider

**Files:**

- Modify:
  `internal/provider/network/dns/debian_update_resolv_conf_by_interface.go` (or
  equivalent write file)
- Modify: corresponding test file

- [ ] **Step 1: Check if DNS uses exec for writes**

Read the DNS provider to determine if it executes commands for writes. The DNS
provider may use `resolvectl` or write files directly. If it uses `RunCmd` for
writes, change to `RunPrivilegedCmd`. If it only writes files via `avfs`, no
changes needed (file writes are covered by capabilities).

- [ ] **Step 2: Update if needed and run tests**

Run: `go test ./internal/provider/network/dns/... -count=1`

- [ ] **Step 3: Commit (if changes made)**

```
refactor(dns): use RunPrivilegedCmd for write operations
```

---

### Task 14: Update documentation

**Files:**

- Modify: `docs/docs/sidebar/usage/configuration.md`
- Create: `docs/docs/sidebar/features/agent-hardening.md`

- [ ] **Step 1: Add config reference to configuration.md**

In the agent section of configuration.md, add:

```yaml
# Least-privilege mode for running the agent as an unprivileged user.
privilege_escalation:
  # Prepend "sudo" to write commands.
  sudo: false
  # Verify Linux capabilities at startup.
  capabilities: false
  # Run privilege checks before accepting jobs.
  preflight: false
```

Add the environment variable mappings:

| Config Key                                | Environment Variable                            |
| ----------------------------------------- | ----------------------------------------------- |
| `agent.privilege_escalation.sudo`         | `OSAPI_AGENT_PRIVILEGE_ESCALATION_SUDO`         |
| `agent.privilege_escalation.capabilities` | `OSAPI_AGENT_PRIVILEGE_ESCALATION_CAPABILITIES` |
| `agent.privilege_escalation.preflight`    | `OSAPI_AGENT_PRIVILEGE_ESCALATION_PREFLIGHT`    |

- [ ] **Step 2: Create agent hardening feature page**

Create `docs/docs/sidebar/features/agent-hardening.md` with:

- Overview of least-privilege mode
- Configuration reference
- Sudoers drop-in (copy from spec)
- Capabilities setup (copy from spec)
- Systemd unit file (copy from spec)
- Preflight output example
- Command privilege reference tables (copy from spec)

- [ ] **Step 3: Update docusaurus.config.ts**

Add "Agent Hardening" to the Features navbar dropdown.

- [ ] **Step 4: Commit**

```
docs: add agent hardening feature page and config reference
```

---

### Task 15: Final verification

- [ ] **Step 1: Run full test suite**

```bash
just go::unit
```

All tests must pass.

- [ ] **Step 2: Build and lint**

```bash
go build ./...
just go::vet
```

- [ ] **Step 3: Verify no stray RunCmd calls for write operations**

Check each write command from the spec against the codebase to confirm it uses
`RunPrivilegedCmd`:

```bash
grep -rn 'RunCmd.*"systemctl".*"start\|stop\|restart\|enable\|disable\|daemon-reload"' \
  internal/provider/ --include="*.go" | grep -v _test.go | grep -v RunPrivilegedCmd
```

Repeat for other write commands (`useradd`, `usermod`, `apt-get`, `shutdown`,
`sysctl -p`, etc.). Expect: no matches.

- [ ] **Step 4: Verify read commands stayed on RunCmd**

```bash
grep -rn 'RunPrivilegedCmd.*"systemctl".*"list-units\|list-unit-files\|show\|is-active\|is-enabled"' \
  internal/provider/ --include="*.go" | grep -v _test.go
```

Expect: no matches (reads should not use RunPrivilegedCmd).
