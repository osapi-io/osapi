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

package client

import (
	"net/http"
	"strings"
)

// Response wraps a domain type with raw JSON for CLI --json mode.
type Response[T any] struct {
	Data    T
	rawJSON []byte
	traceID string
}

// NewResponse creates a Response with the given data and raw JSON body.
func NewResponse[T any](
	data T,
	rawJSON []byte,
) *Response[T] {
	return &Response[T]{
		Data:    data,
		rawJSON: rawJSON,
	}
}

// RawJSON returns the raw HTTP response body.
func (r *Response[T]) RawJSON() []byte {
	return r.rawJSON
}

// TraceID returns the ID of the trace of the request, from the traceresponse
// header of the controller. It is empty when the controller did not trace the
// request. Search the tracing backend for it to see the whole request: the
// controller, the queue, and the work of each agent.
func (r *Response[T]) TraceID() string {
	return r.traceID
}

// newResponse creates a Response, with the trace ID of the HTTP response.
func newResponse[T any](
	data T,
	rawJSON []byte,
	httpResp *http.Response,
) *Response[T] {
	r := NewResponse(data, rawJSON)
	if httpResp != nil {
		r.traceID = traceIDFromHeader(httpResp.Header.Get("traceresponse"))
	}

	return r
}

// traceIDFromHeader returns the trace ID of a W3C traceresponse header,
// version-traceid-spanid-flags, or "" when the header is not in that form.
func traceIDFromHeader(
	header string,
) string {
	parts := strings.Split(header, "-")
	if len(parts) != 4 || len(parts[1]) != 32 {
		return ""
	}

	return parts[1]
}
