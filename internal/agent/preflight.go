// Copyright (c) 2026 John Dewey

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

package agent

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	osexec "os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/exec"
)

// PreflightResult holds the outcome of a single preflight check. A failed
// fatal check stops the agent; any other failed check is a warning, because
// it fails only the operations that need it.
type PreflightResult struct {
	Name   string
	Passed bool
	Error  string
	Fatal  bool
}

// sudoCommands lists binaries the hand-written domains run with sudo. A
// behavior declares its own with behavior.Sudo; requiredSudo adds them.
var sudoCommands = []string{
	"systemctl",
	"sysctl",
	"timedatectl",
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
	"netplan",
	"chpasswd",
}

// requiredCapabilities maps Linux capability names to their bit positions
// in the CapEff bitmask.
var requiredCapabilities = map[string]int{
	"CAP_CHOWN":           0,
	"CAP_DAC_OVERRIDE":    1,
	"CAP_DAC_READ_SEARCH": 2,
	"CAP_FOWNER":          3,
	"CAP_KILL":            5,
	"CAP_NET_ADMIN":       12,
}

// procStatusPath is the path to the proc status file for reading capabilities.
// Overridable in tests via export_test.go.
var procStatusPath = "/proc/self/status"

// RunPreflight checks what privilege escalation needs. It fails only when
// sudo cannot run without a password at all. A program with no sudo rule, or
// a missing capability, is a failed check that is not fatal: it fails the
// operations that need it, and the agent still starts. A program that is not
// installed is not checked, since nothing on the host can run it.
func RunPreflight(
	ctx context.Context,
	logger *slog.Logger,
	execManager exec.Manager,
	platform string,
) ([]PreflightResult, bool) {
	results := checkSudoAccess(ctx, logger, execManager, platform)
	results = append(results, checkCapabilities(logger)...)

	for _, r := range results {
		if r.Fatal && !r.Passed {
			return results, false
		}
	}

	return results, true
}

// checkSudoAccess lists the agent's sudo rules once with sudo -n -l, then
// checks each installed program the agent runs with sudo has a rule. The
// rule's arguments are not checked: sudo -l with a program and no arguments
// fails for a rule that limits them, so a narrow rule such as
// "sysctl -p /etc/sysctl.d/osapi-*.conf" counts as a rule for sysctl.
func checkSudoAccess(
	ctx context.Context,
	logger *slog.Logger,
	execManager exec.Manager,
	platform string,
) []PreflightResult {
	out, err := execManager.RunCmd(ctx, "sudo", []string{"-n", "-l"})
	if err != nil {
		logger.Debug("sudo preflight check failed", slog.String("error", err.Error()))

		return []PreflightResult{{
			Name:   "sudo",
			Passed: false,
			Fatal:  true,
			Error: fmt.Sprintf(
				"sudo -n -l: %s: sudo needs a password or grants this user nothing",
				err.Error(),
			),
		}}
	}

	all, granted := parseSudoList(out)
	results := []PreflightResult{{Name: "sudo", Passed: true, Fatal: true}}

	for _, cmd := range requiredSudo(platform) {
		if _, err := lookPath(cmd); err != nil {
			continue
		}

		name := fmt.Sprintf("sudo:%s", cmd)
		if all || slices.Contains(granted, cmd) {
			results = append(results, PreflightResult{Name: name, Passed: true})

			continue
		}

		results = append(results, PreflightResult{
			Name:   name,
			Passed: false,
			Error:  fmt.Sprintf("no sudo rule runs %s; operations that need it fail", cmd),
		})
	}

	return results
}

// parseSudoList reads the output of sudo -n -l. It returns whether a rule
// grants ALL, and the base name of each program a rule names.
func parseSudoList(
	out string,
) (bool, []string) {
	var granted []string
	inRules := false

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "may run the following commands") {
			inRules = true

			continue
		}
		if !inRules || !strings.HasPrefix(line, " ") {
			continue
		}

		rule := strings.TrimSpace(line)
		// Drop the "(root) NOPASSWD:" prefix: the run-as list, then tags.
		if strings.HasPrefix(rule, "(") {
			if i := strings.Index(rule, ")"); i >= 0 {
				rule = strings.TrimSpace(rule[i+1:])
			}
		}
		for {
			i := strings.Index(rule, ":")
			sp := strings.Index(rule, " ")
			if i < 0 || (sp >= 0 && sp < i) || strings.HasPrefix(rule, "/") {
				break
			}
			rule = strings.TrimSpace(rule[i+1:])
		}

		for _, cmd := range strings.Split(rule, ",") {
			fields := strings.Fields(cmd)
			if len(fields) == 0 {
				continue
			}
			if fields[0] == "ALL" {
				return true, nil
			}
			base := filepath.Base(fields[0])
			if !slices.Contains(granted, base) {
				granted = append(granted, base)
			}
		}
	}

	return false, granted
}

// checkCapabilities verifies that the agent process has the required Linux
// capabilities in its effective capability set.
func checkCapabilities(
	logger *slog.Logger,
) []PreflightResult {
	var results []PreflightResult

	capEff, err := readCapEff()
	if err != nil {
		logger.Debug(
			"capability preflight check failed",
			slog.String("error", err.Error()),
		)

		for capName := range requiredCapabilities {
			results = append(results, PreflightResult{
				Name:   fmt.Sprintf("cap:%s", capName),
				Passed: false,
				Error:  fmt.Sprintf("failed to read capabilities: %s", err.Error()),
			})
		}

		return results
	}

	for capName, bit := range requiredCapabilities {
		name := fmt.Sprintf("cap:%s", capName)

		if capEff&(1<<uint(bit)) != 0 {
			results = append(results, PreflightResult{
				Name:   name,
				Passed: true,
			})
		} else {
			logger.Debug(
				"capability not present",
				slog.String("capability", capName),
				slog.Int("bit", bit),
			)
			results = append(results, PreflightResult{
				Name:   name,
				Passed: false,
				Error:  fmt.Sprintf("capability %s (bit %d) not in effective set", capName, bit),
			})
		}
	}

	return results
}

// readCapEff reads the CapEff line from /proc/self/status and parses it
// as a hexadecimal uint64.
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
			hexStr := strings.TrimSpace(strings.TrimPrefix(line, "CapEff:"))

			val, err := strconv.ParseUint(hexStr, 16, 64)
			if err != nil {
				return 0, fmt.Errorf("parse CapEff %q: %w", hexStr, err)
			}

			return val, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan %s: %w", procStatusPath, err)
	}

	return 0, fmt.Errorf("CapEff not found in %s", procStatusPath)
}

// requiredSudo is every program the agent runs with sudo: the hand-written
// domains', then each registered behavior's, without repeats.
func requiredSudo(
	platform string,
) []string {
	all := slices.Clone(sudoCommands)
	for _, b := range behavior.All() {
		sudo, _ := b.Programs(behavior.PlatformName(platform))
		for _, p := range sudo {
			if !slices.Contains(all, p) {
				all = append(all, p)
			}
		}
	}

	return all
}

// lookPath finds a program on the agent's PATH. Overridable in tests via
// export_test.go.
var lookPath = osexec.LookPath

// MissingPrograms is each program a registered behavior declares with
// behavior.Runs that is not on the agent's PATH, once. A missing one fails
// only the behavior that runs it, so the agent warns rather than refusing to
// start.
func MissingPrograms(
	platform string,
) []string {
	var missing []string
	for _, b := range behavior.All() {
		_, runs := b.Programs(behavior.PlatformName(platform))
		for _, p := range runs {
			// A behavior that runs the program its caller names needs none.
			if p == behavior.AnyProgram || slices.Contains(missing, p) {
				continue
			}
			if _, err := lookPath(p); err != nil {
				missing = append(missing, p)
			}
		}
	}

	return missing
}
