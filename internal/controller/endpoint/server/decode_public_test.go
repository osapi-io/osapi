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

package server_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
)

type decodeInput struct {
	Name  string `json:"name"  validate:"required"`
	Limit int    `json:"limit"`
	Note  string `json:"note"`
}

func (s *ServerPublicTestSuite) TestDecode() {
	tests := []struct {
		name    string
		target  string
		body    string
		wantMsg string
		want    decodeInput
	}{
		{
			name:   "when the path, query and body are given, each sets its field",
			target: "/x/web?limit=3",
			body:   `{"note":"hi"}`,
			want:   decodeInput{Name: "web", Limit: 3, Note: "hi"},
		},
		{
			name:    "when a query parameter does not parse, it says so",
			target:  "/x/web?limit=many",
			wantMsg: "invalid query parameter limit",
		},
		{
			name:    "when the body is not JSON, it says so",
			target:  "/x/web",
			body:    "{",
			wantMsg: "invalid JSON body",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			e := echo.New()
			var got decodeInput
			var msg string
			e.POST("/x/:name", func(c *echo.Context) error {
				var ok bool
				if msg, ok = server.Body(c, &got); !ok {
					return nil
				}
				if msg, ok = server.Param(c, "name", &got.Name); !ok {
					return nil
				}
				if msg, ok = server.Query(c, "limit", &got.Limit); !ok {
					return nil
				}
				msg, _ = server.Valid(got)

				return nil
			})
			req := httptest.NewRequest(http.MethodPost, tc.target, strings.NewReader(tc.body))
			e.ServeHTTP(httptest.NewRecorder(), req)

			if tc.wantMsg != "" {
				s.Contains(msg, tc.wantMsg)

				return
			}
			s.Empty(msg)
			s.Equal(tc.want, got)
		})
	}

	msg, ok := server.Valid(decodeInput{})
	s.False(ok)
	s.NotEmpty(msg)

	type traceInput struct {
		ID    string `json:"id" pattern:"^[0-9a-f]{4}$"`
		Other string
		Count int    `          pattern:"x"`
		Named string `          pattern:"^a$"`
	}
	msg, ok = server.Valid(traceInput{ID: "beef", Named: "a"})
	s.True(ok, msg)
	msg, ok = server.Valid(traceInput{ID: "BEEF", Named: "a"})
	s.False(ok)
	s.Equal("id does not match ^[0-9a-f]{4}$", msg)
	msg, ok = server.Valid(traceInput{ID: "beef", Named: "b"})
	s.False(ok)
	s.Equal("Named does not match ^a$", msg)
}

type uploadForm struct {
	Name string `json:"name"           in:"form"`
	Kind string `json:"kind,omitempty" in:"form"`
	File []byte `json:"file"           in:"form"`
	Note string `json:"note"`
}

// TestForm reads a multipart form into the in:"form" fields of an input: a
// string from its value, a []byte from its file, and nothing into another
// field.
func (s *ServerPublicTestSuite) TestForm() {
	tests := []struct {
		name         string
		body         func() (*bytes.Buffer, string)
		validateFunc func(uploadForm, string, bool)
	}{
		{
			name: "when the form has the fields and the file, each is read",
			body: func() (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				_ = w.WriteField("name", "a.conf")
				_ = w.WriteField("note", "ignored")
				part, _ := w.CreateFormFile("file", "a.conf")
				_, _ = part.Write([]byte("data"))
				_ = w.Close()

				return b, w.FormDataContentType()
			},
			validateFunc: func(in uploadForm, msg string, ok bool) {
				s.Require().True(ok, msg)
				s.Equal(uploadForm{Name: "a.conf", File: []byte("data")}, in)
			},
		},
		{
			name: "when the form has no file, the file is empty",
			body: func() (*bytes.Buffer, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				_ = w.WriteField("name", "a.conf")
				_ = w.Close()

				return b, w.FormDataContentType()
			},
			validateFunc: func(in uploadForm, _ string, ok bool) {
				s.True(ok)
				s.Nil(in.File)
			},
		},
		{
			name: "when the body is not a form, it says so",
			body: func() (*bytes.Buffer, string) {
				return bytes.NewBufferString(`{"name":"a.conf"}`), "application/json"
			},
			validateFunc: func(_ uploadForm, msg string, ok bool) {
				s.False(ok)
				s.Contains(msg, "not a multipart form")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			body, ct := tc.body()
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Set("Content-Type", ct)
			c := echo.New().NewContext(req, httptest.NewRecorder())

			var in uploadForm
			msg, ok := server.Form(c, &in)
			tc.validateFunc(in, msg, ok)
		})
	}
}
