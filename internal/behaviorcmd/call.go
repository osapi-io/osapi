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

package behaviorcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Caller sends one request to the API. The SDK client satisfies it: the SDK
// knows routes, not behaviors, so building a behavior's request is this
// package's work.
type Caller interface {
	Request(
		ctx context.Context,
		method string,
		path string,
		body []byte,
	) (*client.Response[json.RawMessage], error)
}

// Result is what every behavior method answers: the job and one row per
// host, or per element when the method returns a list.
type Result struct {
	JobID   string           `json:"job_id"`
	Results []map[string]any `json:"results"`
}

// Build is the request that runs method m of b on target with input, a value
// of the method's input type: the key in the path, a read's other fields in
// the query, and a write's input as its body.
func Build(
	b *behavior.Behavior,
	m *behavior.Method,
	target string,
	input any,
) (method string, path string, body []byte, err error) {
	fields := map[string]any{}
	if input != nil {
		raw, err := json.Marshal(input)
		if err != nil {
			return "", "", nil, fmt.Errorf("%s: encode input: %w", b.Operation(m), err)
		}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return "", "", nil, fmt.Errorf("%s: input must be an object: %w", b.Operation(m), err)
		}
	}

	path = strings.Replace(b.Route(m), "{hostname}", url.PathEscape(target), 1)
	if m.KeyField != "" {
		path = strings.Replace(
			path,
			"{"+m.KeyField+"}",
			url.PathEscape(fmt.Sprint(fields[m.KeyField])),
			1,
		)
	}

	if m.Verb == behavior.VerbWrite {
		// fields holds decoded JSON, which always encodes.
		body, _ = json.Marshal(fields)
	} else {
		q := url.Values{}
		for k, v := range fields {
			if k != m.KeyField {
				q.Set(k, fmt.Sprint(v))
			}
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}

	return m.HTTPMethod(), path, body, nil
}

// Call sends the request Build makes through c, and returns the decoded
// result and the body as the API sent it.
func Call(
	ctx context.Context,
	c Caller,
	b *behavior.Behavior,
	m *behavior.Method,
	target string,
	input any,
) (Result, []byte, error) {
	method, path, body, err := Build(b, m, target, input)
	if err != nil {
		return Result{}, nil, err
	}

	resp, err := c.Request(ctx, method, path, body)
	if err != nil {
		return Result{}, nil, err
	}

	var out Result
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return Result{}, nil, fmt.Errorf("%s: decode response: %w", b.Operation(m), err)
	}

	return out, resp.RawJSON(), nil
}
