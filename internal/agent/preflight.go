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
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
)

// PreflightResult holds the outcome of a single preflight check.
type PreflightResult struct {
	Name   string
	Passed bool
	Error  string
}

// sudoCommands lists binaries that require sudo access for agent operations.
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

// requiredCapabilities maps Linux capability names to their bit positions
// in the CapEff bitmask.
var requiredCapabilities = map[string]int{
	"CAP_DAC_OVERRIDE":    1,
	"CAP_DAC_READ_SEARCH": 2,
	"CAP_FOWNER":          3,
	"CAP_KILL":            5,
}

// procStatusPath is the path to the proc status file for reading capabilities.
// Overridable in tests via export_test.go.
var procStatusPath = "/proc/self/status"

// RunPreflight runs sudo and capability checks. When called, it always
// checks both sudo access and Linux capabilities. Returns the combined
// results and whether all checks passed.
func RunPreflight(
	logger *slog.Logger,
	execManager exec.Manager,
) ([]PreflightResult, bool) {
	allPassed := true

	sudoResults := checkSudoAccess(logger, execManager)
	capResults := checkCapabilities(logger)

	results := make([]PreflightResult, 0, len(sudoResults)+len(capResults))
	results = append(results, sudoResults...)

	for _, r := range sudoResults {
		if !r.Passed {
			allPassed = false
		}
	}

	results = append(results, capResults...)

	for _, r := range capResults {
		if !r.Passed {
			allPassed = false
		}
	}

	return results, allPassed
}

// checkSudoAccess verifies that the agent can run each required command
// via sudo without a password prompt.
func checkSudoAccess(
	logger *slog.Logger,
	execManager exec.Manager,
) []PreflightResult {
	var results []PreflightResult

	for _, cmd := range sudoCommands {
		name := fmt.Sprintf("sudo:%s", cmd)

		_, err := execManager.RunCmd("sudo", []string{"-n", "which", cmd})
		if err != nil {
			logger.Debug("sudo preflight check failed",
				slog.String("command", cmd),
				slog.String("error", err.Error()),
			)
			results = append(results, PreflightResult{
				Name:   name,
				Passed: false,
				Error:  fmt.Sprintf("sudo -n which %s: %s", cmd, err.Error()),
			})

			continue
		}

		results = append(results, PreflightResult{
			Name:   name,
			Passed: true,
		})
	}

	return results
}

// checkCapabilities verifies that the agent process has the required Linux
// capabilities in its effective capability set.
func checkCapabilities(
	logger *slog.Logger,
) []PreflightResult {
	var results []PreflightResult

	capEff, err := readCapEff()
	if err != nil {
		logger.Debug("capability preflight check failed",
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
			logger.Debug("capability not present",
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
