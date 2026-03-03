# --stdout/--stderr Flags Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `--stdout` and `--stderr` flags to `node command exec` and `node command shell` CLI commands for raw command output without `jq`.

**Architecture:** CLI-only change. Add two boolean flags that bypass the table/JSON display and print raw remote stdout/stderr directly. Multi-host output prefixes each line with a dimmed hostname. Exit code propagates from the remote command.

**Tech Stack:** Go, Cobra, lipgloss (existing `cli.DimStyle`)

---

### Task 1: Add PrintRawOutput helper to internal/cli

**Files:**
- Create: `internal/cli/raw_output.go`
- Create: `internal/cli/raw_output_test.go`

**Step 1: Write the failing test**

Create `internal/cli/raw_output_test.go`:

```go
package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/cli"
)

type RawOutputPublicTestSuite struct {
	suite.Suite
}

func TestRawOutputPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RawOutputPublicTestSuite))
}

func (suite *RawOutputPublicTestSuite) TestPrintRawOutput_SingleHost() {
	tests := []struct {
		name     string
		results  []cli.RawResult
		wantOut  string
		wantErr  string
	}{
		{
			name: "single host stdout only",
			results: []cli.RawResult{
				{Hostname: "server1", Stdout: "file1\nfile2\n", Stderr: ""},
			},
			wantOut: "file1\nfile2\n",
			wantErr: "",
		},
		{
			name: "single host stderr only",
			results: []cli.RawResult{
				{Hostname: "server1", Stdout: "", Stderr: "permission denied\n"},
			},
			wantOut: "",
			wantErr: "permission denied\n",
		},
		{
			name: "single host both",
			results: []cli.RawResult{
				{Hostname: "server1", Stdout: "output\n", Stderr: "warning\n"},
			},
			wantOut: "output\n",
			wantErr: "warning\n",
		},
		{
			name:    "single host empty output",
			results: []cli.RawResult{{Hostname: "server1"}},
			wantOut: "",
			wantErr: "",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var stdout, stderr bytes.Buffer
			cli.PrintRawOutput(&stdout, &stderr, tc.results, true, true)
			assert.Equal(suite.T(), tc.wantOut, stdout.String())
			assert.Equal(suite.T(), tc.wantErr, stderr.String())
		})
	}
}

func (suite *RawOutputPublicTestSuite) TestPrintRawOutput_MultiHost() {
	tests := []struct {
		name     string
		results  []cli.RawResult
		wantOut  string
	}{
		{
			name: "multi host stdout prefixed",
			results: []cli.RawResult{
				{Hostname: "web-01", Stdout: "file1\nfile2\n"},
				{Hostname: "web-02", Stdout: "file3\n"},
			},
			// Hostname prefix is added (without lipgloss color in test)
			wantOut: "web-01  file1\nweb-01  file2\nweb-02  file3\n",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var stdout, stderr bytes.Buffer
			cli.PrintRawOutputPlain(&stdout, &stderr, tc.results, true, false)
			assert.Equal(suite.T(), tc.wantOut, stdout.String())
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestRawOutputPublicTestSuite -v ./internal/cli/...`
Expected: FAIL — `PrintRawOutput`, `PrintRawOutputPlain`, `RawResult` not defined

**Step 3: Write minimal implementation**

Create `internal/cli/raw_output.go`:

```go
package cli

import (
	"fmt"
	"io"
	"strings"
)

// RawResult holds raw command output for a single host.
type RawResult struct {
	Hostname string
	Stdout   string
	Stderr   string
	ExitCode int
}

// PrintRawOutput writes raw command output to the given writers.
// For single results, output is printed without hostname prefix.
// For multiple results, each line is prefixed with a dimmed hostname.
// showStdout/showStderr control which streams are printed.
func PrintRawOutput(
	stdout io.Writer,
	stderr io.Writer,
	results []RawResult,
	showStdout bool,
	showStderr bool,
) {
	multiHost := len(results) > 1

	for _, r := range results {
		if showStdout && r.Stdout != "" {
			writeLines(stdout, r.Hostname, r.Stdout, multiHost, true)
		}
		if showStderr && r.Stderr != "" {
			writeLines(stderr, r.Hostname, r.Stderr, multiHost, true)
		}
	}
}

// PrintRawOutputPlain writes raw output without lipgloss styling.
// Used for testing and non-TTY output.
func PrintRawOutputPlain(
	stdout io.Writer,
	stderr io.Writer,
	results []RawResult,
	showStdout bool,
	showStderr bool,
) {
	multiHost := len(results) > 1

	for _, r := range results {
		if showStdout && r.Stdout != "" {
			writeLines(stdout, r.Hostname, r.Stdout, multiHost, false)
		}
		if showStderr && r.Stderr != "" {
			writeLines(stderr, r.Hostname, r.Stderr, multiHost, false)
		}
	}
}

func writeLines(
	w io.Writer,
	hostname string,
	content string,
	multiHost bool,
	styled bool,
) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line == "" && content[len(content)-1] == '\n' {
			continue
		}
		if multiHost {
			prefix := hostname
			if styled {
				prefix = DimStyle.Render(hostname)
			}
			fmt.Fprintf(w, "%s  %s\n", prefix, line)
		} else {
			fmt.Fprintln(w, line)
		}
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestRawOutputPublicTestSuite -v ./internal/cli/...`
Expected: PASS

**Step 5: Commit**

```
git add internal/cli/raw_output.go internal/cli/raw_output_test.go
git commit -m "feat(cli): add PrintRawOutput helper for --stdout/--stderr flags"
```

---

### Task 2: Add --stdout/--stderr flags to command exec

**Files:**
- Modify: `cmd/client_node_command_exec.go`

**Step 1: Add flag definitions in init()**

Add after line 132 (`Int("timeout", ...)`):

```go
	clientNodeCommandExecCmd.PersistentFlags().
		Bool("stdout", false, "Print only remote stdout")
	clientNodeCommandExecCmd.PersistentFlags().
		Bool("stderr", false, "Print only remote stderr")
```

**Step 2: Add raw output handling in the Run function**

Read the new flags after `timeout` (after line 43):

```go
		showStdout, _ := cmd.Flags().GetBool("stdout")
		showStderr, _ := cmd.Flags().GetBool("stderr")
```

Replace the `case http.StatusAccepted:` block (lines 66-99) with:

```go
		case http.StatusAccepted:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if (showStdout || showStderr) && resp.JSON202 != nil {
				results := make([]cli.RawResult, 0, len(resp.JSON202.Results))
				maxExitCode := 0
				for _, r := range resp.JSON202.Results {
					exitCode := 0
					if r.ExitCode != nil {
						exitCode = *r.ExitCode
					}
					if exitCode > maxExitCode {
						maxExitCode = exitCode
					}
					results = append(results, cli.RawResult{
						Hostname: r.Hostname,
						Stdout:   cli.SafeString(r.Stdout),
						Stderr:   cli.SafeString(r.Stderr),
						ExitCode: exitCode,
					})
				}
				cli.PrintRawOutput(os.Stdout, os.Stderr, results, showStdout, showStderr)
				if maxExitCode != 0 {
					os.Exit(maxExitCode)
				}
				return
			}

			if resp.JSON202 != nil && resp.JSON202.JobId != nil {
				fmt.Println()
				cli.PrintKV("Job ID", resp.JSON202.JobId.String())
			}

			if resp.JSON202 != nil && len(resp.JSON202.Results) > 0 {
				results := make([]cli.ResultRow, 0, len(resp.JSON202.Results))
				for _, r := range resp.JSON202.Results {
					results = append(results, cli.ResultRow{
						Hostname: r.Hostname,
						Changed:  r.Changed,
						Error:    r.Error,
						Fields: []string{
							cli.SafeString(r.Stdout),
							cli.SafeString(r.Stderr),
							cli.IntToSafeString(r.ExitCode),
							formatDurationMs(r.DurationMs),
						},
					})
				}
				headers, rows := cli.BuildBroadcastTable(results, []string{
					"STDOUT",
					"STDERR",
					"EXIT CODE",
					"DURATION",
				})
				cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
			}
```

Add `"os"` to the imports.

**Step 3: Run tests and build**

Run: `go build ./... && just go::unit`
Expected: PASS

**Step 4: Commit**

```
git add cmd/client_node_command_exec.go
git commit -m "feat(cli): add --stdout/--stderr flags to node command exec"
```

---

### Task 3: Add --stdout/--stderr flags to command shell

**Files:**
- Modify: `cmd/client_node_command_shell.go`

**Step 1: Add flag definitions in init()**

Add after line 118 (`Int("timeout", ...)`):

```go
	clientNodeCommandShellCmd.PersistentFlags().
		Bool("stdout", false, "Print only remote stdout")
	clientNodeCommandShellCmd.PersistentFlags().
		Bool("stderr", false, "Print only remote stderr")
```

**Step 2: Add raw output handling in the Run function**

Read the new flags after `timeout` (after line 41):

```go
		showStdout, _ := cmd.Flags().GetBool("stdout")
		showStderr, _ := cmd.Flags().GetBool("stderr")
```

Replace the `case http.StatusAccepted:` block (lines 62-96) with the
same pattern as Task 2 (identical logic).

Add `"os"` to the imports.

**Step 3: Run tests and build**

Run: `go build ./... && just go::unit`
Expected: PASS

**Step 4: Commit**

```
git add cmd/client_node_command_shell.go
git commit -m "feat(cli): add --stdout/--stderr flags to node command shell"
```

---

### Task 4: Update CLI documentation

**Files:**
- Modify: `docs/docs/sidebar/usage/cli/client/node/command/exec.md`
- Modify: `docs/docs/sidebar/usage/cli/client/node/command/shell.md`

**Step 1: Update exec.md**

Add a new section after "## JSON Output" and before "## Flags":

```markdown
## Raw Output

Use `--stdout` to print only the remote command's stdout, without the
table wrapper:

```bash
$ osapi client node command exec --command ls --args "-la" --stdout
total 48
drwxr-xr-x  12 john  staff  384 Mar  2 10:00 .
-rw-r--r--   1 john  staff 1234 Mar  2 09:30 main.go
```

Use `--stderr` to print only stderr:

```bash
$ osapi client node command exec --command ls --args "/nonexistent" --stderr
ls: cannot access '/nonexistent': No such file or directory
```

Both flags can be combined. When targeting multiple hosts, each line is
prefixed with the hostname:

```bash
$ osapi client node command exec --command hostname --target _all --stdout
  web-01  web-01.example.com
  web-02  web-02.example.com
```

The CLI exit code matches the remote command's exit code, making it
scriptable:

```bash
$ osapi client node command exec --command "test" --args "-f,/etc/hosts" --stdout && echo exists
exists
```
```

Add the new flags to the Flags table:

```markdown
| `--stdout`     | Print only remote stdout                                 |         |
| `--stderr`     | Print only remote stderr                                 |         |
```

**Step 2: Update shell.md**

Add the same "## Raw Output" section with shell-appropriate examples:

```markdown
## Raw Output

Use `--stdout` to print only the remote command's stdout:

```bash
$ osapi client node command shell --command "df -h / | tail -1" --stdout
/dev/sda1        50G   12G   35G  26% /
```

Use `--stderr` to print only stderr:

```bash
$ osapi client node command shell --command "cat /nonexistent" --stderr
cat: /nonexistent: No such file or directory
```

Both flags can be combined. When targeting multiple hosts, each line is
prefixed with the hostname:

```bash
$ osapi client node command shell --command "uname -r" --target _all --stdout
  web-01  5.15.0-91-generic
  web-02  5.15.0-91-generic
```

The CLI exit code matches the remote command's exit code.
```

Add the new flags to the Flags table:

```markdown
| `--stdout`     | Print only remote stdout                                 |         |
| `--stderr`     | Print only remote stderr                                 |         |
```

**Step 3: Verify docs build**

Run: `just docs::fmt-check` (or `just docs::build` if available)

**Step 4: Commit**

```
git add docs/docs/sidebar/usage/cli/client/node/command/exec.md
git add docs/docs/sidebar/usage/cli/client/node/command/shell.md
git commit -m "docs: add --stdout/--stderr flag documentation for exec and shell"
```

---

### Task 5: Final verification

**Step 1: Run full test suite**

Run: `just test`
Expected: All lint + unit tests pass

**Step 2: Build and smoke test**

Run: `go build -o osapi . && ./osapi client node command exec --help`
Expected: `--stdout` and `--stderr` flags visible in help output

**Step 3: Commit any fixes**

If anything needed fixing, commit with appropriate message.
