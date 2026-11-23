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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	return c.Send(ctx, method, path, "application/json", body)
}

// Send is Request with the content type of the body, such as the
// multipart/form-data of an upload.
func (c *Client) Send(
	ctx context.Context,
	method string,
	path string,
	contentType string,
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
		req.Header.Set("Content-Type", contentType)
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
		var e ErrorResponse
		_ = json.Unmarshal(raw, &e)
		err := checkError(resp.StatusCode, &e)
		var withBody interface{ setBody([]byte) }
		if errors.As(err, &withBody) {
			withBody.setBody(raw)
		}

		return nil, err
	}

	return newResponse(json.RawMessage(raw), raw, resp), nil
}

// send is how each typed method sends its request: to path, with query when
// it has parameters, and body as JSON when it is not nil. It decodes the
// answer's body into T. An answer outside 2xx is the SDK's error types, from
// Send.
func send[T any](
	ctx context.Context,
	c *Client,
	method string,
	path string,
	query url.Values,
	body any,
) (*Response[T], error) {
	resp, err := sendJSON(ctx, c, method, path, query, body)
	if err != nil {
		return nil, err
	}

	return decode[T](method, path, resp)
}

// decode is the answer resp of method and path, with its body decoded
// into T.
func decode[T any](
	method string,
	path string,
	resp *Response[json.RawMessage],
) (*Response[T], error) {
	var data T
	if err := json.Unmarshal(resp.rawJSON, &data); err != nil {
		return nil, fmt.Errorf("%s %s: decode response: %w", method, path, err)
	}
	out := NewResponse(data, resp.rawJSON)
	out.traceID = resp.traceID

	return out, nil
}

// sendEmpty is send for a method whose answer has no body, such as a 204.
func sendEmpty(
	ctx context.Context,
	c *Client,
	method string,
	path string,
	query url.Values,
	body any,
) error {
	_, err := sendJSON(ctx, c, method, path, query, body)

	return err
}

// sendJSON is Send with query on the path and body encoded as JSON.
func sendJSON(
	ctx context.Context,
	c *Client,
	method string,
	path string,
	query url.Values,
	body any,
) (*Response[json.RawMessage], error) {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var data []byte
	if body != nil {
		var err error
		if data, err = json.Marshal(body); err != nil {
			return nil, fmt.Errorf("%s %s: encode request: %w", method, path, err)
		}
	}

	return c.Send(ctx, method, path, "application/json", data)
}
