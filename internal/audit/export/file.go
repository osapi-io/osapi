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

package export

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	gen "github.com/retr0h/osapi/internal/client/gen"
)

// OpenFileFunc returns the current file opener for test inspection.
func OpenFileFunc() func(string) (io.WriteCloser, error) {
	return openFile
}

// SetOpenFileFunc replaces the file opener. Used by tests.
func SetOpenFileFunc(
	fn func(string) (io.WriteCloser, error),
) {
	openFile = fn
}

// openFile is the function used to open files. Override in tests.
var openFile = func(
	path string,
) (io.WriteCloser, error) {
	return os.Create(path)
}

// FileExporter writes audit entries as JSON lines to a file.
type FileExporter struct {
	Path   string
	file   io.WriteCloser
	writer *bufio.Writer
}

// NewFileExporter creates a new FileExporter for the given path.
func NewFileExporter(
	path string,
) *FileExporter {
	return &FileExporter{
		Path: path,
	}
}

// Open creates the output file and prepares for writing.
func (e *FileExporter) Open(
	_ context.Context,
) error {
	f, err := openFile(e.Path)
	if err != nil {
		return fmt.Errorf("opening export file: %w", err)
	}

	e.file = f
	e.writer = bufio.NewWriter(f)

	return nil
}

// Write marshals an audit entry to JSON and writes it as a single line.
func (e *FileExporter) Write(
	_ context.Context,
	entry gen.AuditEntry,
) error {
	if e.writer == nil {
		return fmt.Errorf("exporter not opened")
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling entry: %w", err)
	}

	if _, err := e.writer.Write(data); err != nil {
		return fmt.Errorf("writing entry: %w", err)
	}

	if err := e.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("writing newline: %w", err)
	}

	return nil
}

// Close flushes the buffer and closes the file.
func (e *FileExporter) Close(
	_ context.Context,
) error {
	if e.writer == nil {
		return fmt.Errorf("exporter not opened")
	}

	if err := e.writer.Flush(); err != nil {
		return fmt.Errorf("flushing writer: %w", err)
	}

	if err := e.file.Close(); err != nil {
		return fmt.Errorf("closing file: %w", err)
	}

	return nil
}
