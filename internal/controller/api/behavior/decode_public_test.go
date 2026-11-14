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

package behavior_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	apibehavior "github.com/osapi-io/osapi/internal/controller/api/behavior"
)

type DecodePublicTestSuite struct {
	suite.Suite
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("broken body") }

// mode is a named string type, as an input field can be.
type mode string

// newContext is an echo context for a request to target with body, and the
// path parameter key set to keyValue.
func newContext(
	target string,
	body io.Reader,
	keyValue string,
) *echo.Context {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, target, body), httptest.NewRecorder())
	c.SetPathValues(echo.PathValues{{Name: "key", Value: keyValue}})

	return c
}

func (s *DecodePublicTestSuite) TestBody() {
	tests := []struct {
		name    string
		body    io.Reader
		want    kvIn
		wantMsg string
	}{
		{
			name: "when the body is JSON, it fills the input",
			body: strings.NewReader(`{"key":"k","count":3}`),
			want: kvIn{Key: "k", Count: 3},
		},
		{
			name: "when there is no body, the input is left as it is",
			body: http.NoBody,
		},
		{
			name:    "when the body is not JSON, it says so",
			body:    strings.NewReader("{"),
			wantMsg: "invalid JSON body",
		},
		{
			name:    "when a field has the wrong type, it says so",
			body:    strings.NewReader(`{"count":"three"}`),
			wantMsg: "invalid JSON body",
		},
		{
			name:    "when the body cannot be read, it says so",
			body:    failingBody{},
			wantMsg: "read body: broken body",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var in kvIn
			msg, ok := apibehavior.Body(newContext("/", tc.body, ""), &in)

			if tc.wantMsg != "" {
				s.False(ok)
				s.Contains(msg, tc.wantMsg)

				return
			}
			s.True(ok)
			s.Equal(tc.want, in)
		})
	}
}

func (s *DecodePublicTestSuite) TestQuery() {
	tests := []struct {
		name      string
		target    string
		wantCount int
		wantMode  mode
		wantMsg   string
	}{
		{
			name:   "when the request has neither, both are left as they are",
			target: "/",
		},
		{
			name:      "when a number spells its value, it is decoded",
			target:    "/?count=3&mode=fast",
			wantCount: 3,
			wantMode:  "fast",
		},
		{
			name:    "when a value is not the field's type, it says so",
			target:  "/?count=three",
			wantMsg: "invalid query parameter count",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := newContext(tc.target, http.NoBody, "")
			var count int
			var m mode

			msg, ok := apibehavior.Query(c, "count", &count)
			if tc.wantMsg != "" {
				s.False(ok)
				s.Contains(msg, tc.wantMsg)

				return
			}
			s.True(ok)
			_, ok = apibehavior.Query(c, "mode", &m)
			s.True(ok)
			s.Equal(tc.wantCount, count)
			s.Equal(tc.wantMode, m)
		})
	}
}

func (s *DecodePublicTestSuite) TestParam() {
	tests := []struct {
		name    string
		value   string
		want    int
		wantMsg string
	}{
		{
			name:  "when the key spells the field's type, it is decoded",
			value: "7",
			want:  7,
		},
		{
			name:    "when the key is not the field's type, it says so",
			value:   "seven",
			wantMsg: "invalid path parameter key",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var key int
			msg, ok := apibehavior.Param(newContext("/", http.NoBody, tc.value), "key", &key)

			if tc.wantMsg != "" {
				s.False(ok)
				s.Contains(msg, tc.wantMsg)

				return
			}
			s.True(ok)
			s.Equal(tc.want, key)
		})
	}
}

func (s *DecodePublicTestSuite) TestValid() {
	tests := []struct {
		name   string
		in     kvIn
		wantOK bool
	}{
		{
			name:   "when the input meets its tags",
			in:     kvIn{Key: "k"},
			wantOK: true,
		},
		{
			name: "when a required field is missing",
			in:   kvIn{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msg, ok := apibehavior.Valid(tc.in)

			s.Equal(tc.wantOK, ok)
			s.Equal(tc.wantOK, msg == "")
		})
	}
}

func TestDecodePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DecodePublicTestSuite))
}
