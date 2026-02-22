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

package worker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/command"
)

// processCommandOperation handles command-related operations.
func (w *Worker) processCommandOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract base operation from dotted operation (e.g., "exec.execute" -> "exec")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "exec":
		return w.processCommandExec(jobRequest)
	case "shell":
		return w.processCommandShell(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported command operation: %s", jobRequest.Operation)
	}
}

// processCommandExec handles direct command execution.
func (w *Worker) processCommandExec(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var execData job.CommandExecData
	if err := json.Unmarshal(jobRequest.Data, &execData); err != nil {
		return nil, fmt.Errorf("failed to parse command exec data: %w", err)
	}

	commandProvider := w.getCommandProvider()
	result, err := commandProvider.Exec(command.ExecParams{
		Command: execData.Command,
		Args:    execData.Args,
		Cwd:     execData.Cwd,
		Timeout: execData.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("command exec failed: %w", err)
	}

	return json.Marshal(result)
}

// processCommandShell handles shell command execution.
func (w *Worker) processCommandShell(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var shellData job.CommandShellData
	if err := json.Unmarshal(jobRequest.Data, &shellData); err != nil {
		return nil, fmt.Errorf("failed to parse command shell data: %w", err)
	}

	commandProvider := w.getCommandProvider()
	result, err := commandProvider.Shell(command.ShellParams{
		Command: shellData.Command,
		Cwd:     shellData.Cwd,
		Timeout: shellData.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("command shell failed: %w", err)
	}

	return json.Marshal(result)
}
