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

package audit_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/audit"
)

// SummaryPublicTestSuite covers what an audit entry keeps of a request: enough to
// answer what ran, never a secret, never without bound.
type SummaryPublicTestSuite struct {
	suite.Suite
}

func (s *SummaryPublicTestSuite) TestSummarize() {
	tests := []struct {
		name         string
		body         string
		contentType  string
		validateFunc func(string)
	}{
		{
			name:        "a command and its arguments, which is the whole point",
			body:        `{"command":"systemctl","args":["restart","nginx"]}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "systemctl")
				s.Contains(got, "restart")
			},
		},
		{
			name:        "a password is not recorded",
			body:        `{"name":"deploy","password":"hunter2"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "hunter2")
				s.Contains(got, "[redacted]")
				s.Contains(got, "deploy")
			},
		},
		{
			name:        "a secret nested inside a list is not recorded either",
			body:        `{"users":[{"name":"a","password_hash":"$6$abc"}]}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "$6$abc")
				s.Contains(got, "[redacted]")
			},
		},
		{
			name:        "field names match whatever their case",
			body:        `{"Token":"abc123","Secret":"xyz"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.NotContains(got, "abc123")
				s.NotContains(got, "xyz")
			},
		},
		{
			name:        "an empty body summarizes to nothing",
			body:        "",
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Empty(got)
			},
		},
		{
			name:        "a body that is not json is described, not stored",
			body:        "--boundary\r\nContent-Disposition: form-data\r\n\r\nbinary",
			contentType: "multipart/form-data; boundary=boundary",
			validateFunc: func(got string) {
				s.Contains(got, "multipart/form-data")
				s.NotContains(got, "binary")
			},
		},
		{
			name:        "a body with no content type at all",
			body:        "something",
			contentType: "",
			validateFunc: func(got string) {
				s.Contains(got, "no content type")
			},
		},
		{
			name:        "json that does not parse",
			body:        `{"command":`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "unparsable json")
			},
		},
		{
			name:        "a long value is cut and says so",
			body:        `{"content_note":"` + strings.Repeat("x", 4000) + `"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Less(len(got), 2200)
				s.Contains(got, "[truncated]")
			},
		},
		{
			name:        "a cut that would land inside a character moves back",
			body:        `{"note":"` + strings.Repeat("é", 2000) + `"}`,
			contentType: "application/json",
			validateFunc: func(got string) {
				s.Contains(got, "[truncated]")
				s.True(utf8.ValidString(got), "a truncated summary is still text")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(audit.Summarize([]byte(tt.body), tt.contentType))
		})
	}
}

// TestSummarizeUnencodable covers the encoder failing, which a body that parsed
// cannot cause on its own.
func (s *SummaryPublicTestSuite) TestSummarizeUnencodable() {
	defer audit.ResetMarshalJSON()

	audit.SetMarshalJSON(func(any) ([]byte, error) {
		return nil, errors.New("encoder failed")
	})

	got := audit.Summarize([]byte(`{"command":"ls"}`), "application/json")
	s.Contains(got, "unsummarizable json")
	s.NotContains(got, "ls")
}

// TestJobIDSink covers the channel by which an entry learns its job ID without
// every handler reporting one.
func (s *SummaryPublicTestSuite) TestJobIDSink() {
	s.Run("a context with no sink records nothing and does not panic", func() {
		ctx := context.Background()
		audit.RecordJobID(ctx, "job-1")
		s.Empty(audit.JobIDFrom(ctx))
	})

	s.Run("the job a request created", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "job-1")
		s.Equal("job-1", audit.JobIDFrom(ctx))
	})

	s.Run("the first job wins", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "job-1")
		audit.RecordJobID(ctx, "job-2")
		s.Equal("job-1", audit.JobIDFrom(ctx))
	})

	s.Run("an empty job id is not a job", func() {
		ctx := audit.WithJobIDSink(context.Background())
		audit.RecordJobID(ctx, "")
		s.Empty(audit.JobIDFrom(ctx))
	})

	s.Run("concurrent dispatches do not race", func() {
		ctx := audit.WithJobIDSink(context.Background())

		var wg sync.WaitGroup

		wg.Add(4)

		for i := range 4 {
			go func(n int) {
				defer wg.Done()

				audit.RecordJobID(ctx, string(rune('a'+n)))
				_ = audit.JobIDFrom(ctx)
			}(i)
		}

		wg.Wait()

		s.NotEmpty(audit.JobIDFrom(ctx))
	})
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestSummaryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SummaryPublicTestSuite))
}
