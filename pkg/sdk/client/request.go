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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// Request sends one request to an API path, such as
// /api/node/web-01/sysctl/vm.swappiness, with body as its JSON body when body
// is not nil, and returns the response body. It is for a route the SDK has no
// typed method for. An answer outside 2xx is returned as the SDK's error
// types, as every typed method returns it.
func (c *Client) Request(
	ctx context.Context,
	method string,
	path string,
	body []byte,
) (*Response[json.RawMessage], error) {
	what := method + " " + path

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		strings.TrimSuffix(c.baseURL, "/")+path,
		reader,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", what, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", what, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", what, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e gen.ErrorResponse
		_ = json.Unmarshal(raw, &e)

		return nil, checkError(resp.StatusCode, &e)
	}

	return NewResponse(json.RawMessage(raw), raw), nil
}
