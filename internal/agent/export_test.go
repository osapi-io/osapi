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
	"io/fs"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// SetEmbeddedFS overrides the embedded filesystem for testing.
func SetEmbeddedFS(f fs.FS) {
	embeddedFS = f
}

// ResetEmbeddedFS restores the default embedded filesystem.
func ResetEmbeddedFS() {
	embeddedFS = systemTemplates
}

// SetReadEmbeddedFile overrides the read function for testing.
func SetReadEmbeddedFile(fn func(string) ([]byte, error)) {
	readEmbeddedFile = fn
}

// ResetReadEmbeddedFile restores the default read function.
func ResetReadEmbeddedFile() {
	readEmbeddedFile = func(path string) ([]byte, error) {
		return systemTemplates.ReadFile(path)
	}
}

// ExportProcessScheduleOperation exposes the private processScheduleOperation method for testing.
func ExportProcessScheduleOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processScheduleOperation(req)
}

// ExportProcessCronOperation exposes the private processCronOperation method for testing.
func ExportProcessCronOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processCronOperation(req)
}

// ExportGetCronProvider exposes the private getCronProvider method for testing.
func ExportGetCronProvider(
	a *Agent,
) cron.Provider {
	return a.getCronProvider()
}

// ExportProcessJobOperation exposes the private processJobOperation method for testing.
func ExportProcessJobOperation(
	a *Agent,
	req job.Request,
) (json.RawMessage, error) {
	return a.processJobOperation(req)
}
