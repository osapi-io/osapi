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
package api_test

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
)

type ErrorsPublicTestSuite struct {
	suite.Suite
}

// TestErrorBody holds each error that a route returns to the body of each
// error of the API, {"error": "..."}.
func (s *ErrorsPublicTestSuite) TestErrorBody() {
	tests := []struct {
		name     string
		method   string
		route    echo.HandlerFunc
		wantCode int
		wantBody string
	}{
		{
			name:   "when a parameter does not bind, the 400 says why in error",
			method: http.MethodGet,
			route: func(_ *echo.Context) error {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					"query parameter 'path' is required",
				)
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"query parameter 'path' is required"}`,
		},
		{
			name:     "when the route is not there, the 404 is in error",
			method:   http.MethodGet,
			wantCode: http.StatusNotFound,
			wantBody: `{"error":"Not Found"}`,
		},
		{
			name:   "when a route fails with a plain error, it is a 500 that tells no detail",
			method: http.MethodGet,
			route: func(_ *echo.Context) error {
				return errors.New("dial tcp 10.0.0.9:4222: refused")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"Internal Server Error"}`,
		},
		{
			name:   "when a 500 has a message, it still tells only its status",
			method: http.MethodGet,
			route: func(_ *echo.Context) error {
				return echo.NewHTTPError(http.StatusInternalServerError, "nats: timeout")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"Internal Server Error"}`,
		},
		{
			name:   "when the request is a HEAD, the error has no body",
			method: http.MethodHead,
			route: func(_ *echo.Context) error {
				return echo.NewHTTPError(http.StatusBadRequest, "bad")
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when the route already answered, the error changes nothing",
			method: http.MethodGet,
			route: func(c *echo.Context) error {
				_ = c.NoContent(http.StatusAccepted)

				return errors.New("after the answer")
			},
			wantCode: http.StatusAccepted,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			srv := api.New(config.Config{}, slog.Default())
			if tc.route != nil {
				srv.Echo.Add(tc.method, "/api/thing", tc.route)
			}
			path := "/api/thing"
			if tc.route == nil {
				path = "/api/nothing-here"
			}
			rec := httptest.NewRecorder()
			srv.Echo.ServeHTTP(rec, httptest.NewRequest(tc.method, path, nil))

			s.Equal(tc.wantCode, rec.Code)
			if tc.wantBody == "" {
				s.Empty(rec.Body.String())

				return
			}
			s.JSONEq(tc.wantBody, rec.Body.String())
		})
	}
}

func TestErrorsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ErrorsPublicTestSuite))
}
