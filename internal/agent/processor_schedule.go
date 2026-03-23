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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// processScheduleOperation handles schedule-related operations.
func (a *Agent) processScheduleOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	if a.cronProvider == nil {
		return nil, fmt.Errorf("cron provider not available")
	}

	// Extract base operation from dotted operation (e.g., "cron.list.get" -> "cron")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "cron":
		return a.processCronOperation(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported schedule operation: %s", jobRequest.Operation)
	}
}

// processCronOperation dispatches cron sub-operations.
func (a *Agent) processCronOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract sub-operation: "cron.list.get" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid cron operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	switch subOp {
	case "list":
		return a.processCronList()
	case "get":
		return a.processCronGet(jobRequest)
	case "create":
		return a.processCronCreate(jobRequest)
	case "update":
		return a.processCronUpdate(jobRequest)
	case "delete":
		return a.processCronDelete(jobRequest)
	default:
		return nil, fmt.Errorf("unsupported cron operation: %s", jobRequest.Operation)
	}
}

// processCronList lists all cron entries.
func (a *Agent) processCronList() (json.RawMessage, error) {
	a.logger.Debug("executing cron.List")

	entries, err := a.cronProvider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list cron entries: %w", err)
	}

	return json.Marshal(entries)
}

// processCronGet gets a single cron entry by name.
func (a *Agent) processCronGet(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal cron get data: %w", err)
	}

	a.logger.Debug("executing cron.Get",
		"name", data.Name,
	)

	entry, err := a.cronProvider.Get(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron entry: %w", err)
	}

	return json.Marshal(entry)
}

// processCronCreate creates a new cron entry.
func (a *Agent) processCronCreate(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry cron.CronEntry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cron create data: %w", err)
	}

	a.logger.Debug("executing cron.Create",
		"name", entry.Name,
	)

	result, err := a.cronProvider.Create(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron entry: %w", err)
	}

	return json.Marshal(result)
}

// processCronUpdate updates an existing cron entry.
func (a *Agent) processCronUpdate(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry cron.CronEntry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cron update data: %w", err)
	}

	a.logger.Debug("executing cron.Update",
		"name", entry.Name,
	)

	result, err := a.cronProvider.Update(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to update cron entry: %w", err)
	}

	return json.Marshal(result)
}

// processCronDelete deletes a cron entry.
func (a *Agent) processCronDelete(
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal cron delete data: %w", err)
	}

	a.logger.Debug("executing cron.Delete",
		"name", data.Name,
	)

	result, err := a.cronProvider.Delete(data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete cron entry: %w", err)
	}

	return json.Marshal(result)
}

// getCronProvider returns the injected cron provider.
func (a *Agent) getCronProvider() cron.Provider {
	return a.cronProvider
}
