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
	"log/slog"
	"net/http"
)

// ExportCheckError exposes the private checkError function for testing.
func ExportCheckError(
	statusCode int,
	body *ErrorResponse,
) error {
	return checkError(statusCode, body)
}

// ExportNewAuthTransport exposes the private authTransport constructor for
// testing.
func ExportNewAuthTransport(
	base http.RoundTripper,
	authHeader string,
	logger *slog.Logger,
) http.RoundTripper {
	return &authTransport{
		base:       base,
		authHeader: authHeader,
		logger:     logger,
	}
}

// ExportBaseTransport returns the base transport of the client for testing.
func ExportBaseTransport(
	c *Client,
) http.RoundTripper {
	return c.baseTransport
}

// ExportCloneTransport exposes the private cloneTransport for testing.
func ExportCloneTransport(
	base http.RoundTripper,
) http.RoundTripper {
	return cloneTransport(base)
}

// ExportNewResponse exposes newResponse for testing.
func ExportNewResponse[T any](
	data T,
	rawJSON []byte,
	httpResp *http.Response,
) *Response[T] {
	return newResponse(data, rawJSON, httpResp)
}

// ExportSendJSON exposes the private sendJSON for testing.
var ExportSendJSON = sendJSON
