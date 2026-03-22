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

package metrics

import (
	"fmt"
	"net/http"
)

// handleHealth handles GET /health and always returns 200 with {"status":"ok"}.
func (s *Server) handleHealth(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ok"}`)
}

// handleReady handles GET /health/ready. It returns 503 when no readiness
// func is configured or the readiness func returns an error, and 200 when
// the readiness func returns nil.
func (s *Server) handleReady(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if s.readinessFunc == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"status":"not_ready","error":"readiness check not configured"}`)

		return
	}

	if err := s.readinessFunc(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, `{"status":"not_ready","error":%q}`, err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ready"}`)
}
