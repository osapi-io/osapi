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
	"testing"
	"time"

	"github.com/google/uuid"
	gen "github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
	"github.com/stretchr/testify/suite"
)

type FileInternalTestSuite struct {
	suite.Suite
}

func (s *FileInternalTestSuite) TestWriteNewlineError() {
	entry := gen.AuditEntry{
		Id:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Timestamp:    time.Date(2026, 2, 21, 10, 30, 0, 0, time.UTC),
		User:         "user@example.com",
		Roles:        []string{"admin"},
		Method:       "GET",
		Path:         "/system/hostname",
		SourceIp:     "127.0.0.1",
		ResponseCode: 200,
		DurationMs:   42,
	}

	tests := []struct {
		name         string
		validateFunc func(err error)
	}{
		{
			name: "when WriteByte triggers flush failure returns newline error",
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "writing newline")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Marshal to get the exact byte size of the JSON data.
			data, err := json.Marshal(entry)
			s.Require().NoError(err)

			// Create a bufio.Writer with a buffer sized exactly to the
			// JSON data. Write(data) fills the buffer completely, then
			// WriteByte('\n') finds Available()==0 and calls Flush(),
			// which hits the failing underlying writer.
			fw := &internalFailWriter{}
			e := &FileExporter{
				writer: bufio.NewWriterSize(fw, len(data)),
			}

			err = e.Write(context.Background(), entry)
			tt.validateFunc(err)
		})
	}
}

func TestFileInternalTestSuite(t *testing.T) {
	suite.Run(t, new(FileInternalTestSuite))
}

// internalFailWriter always returns an error on Write.
type internalFailWriter struct{}

func (w *internalFailWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("write failed")
}
