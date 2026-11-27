//go:build integration

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

package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// specProxy sits between the CLI and the server. It checks each request to
// /api/ and its response against the full spec, which osapi generates, and
// forwards both: the spec must still describe what the server does. Each
// request or response that breaks it is a violation, which fails the run.
type specProxy struct {
	router     routers.Router
	mu         sync.Mutex
	violations []string
	// listeners are the proxy's own, which Close closes.
	listeners []net.Listener
}

// Close stops the proxy's listeners.
func (p *specProxy) Close() {
	for _, l := range p.listeners {
		_ = l.Close()
	}
}

// startSpecProxy listens on a free port, forwards to the API at apiURL, and
// is the port the CLI calls.
func startSpecProxy(
	apiURL string,
) (*specProxy, int, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(
		filepath.Join(repoRoot(), "internal", "controller", "api", "gen", "api.yaml"),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("load the full spec: %w", err)
	}
	// The spec names the server's paths, not this host and port.
	doc.Servers = nil
	router, err := legacy.NewRouter(doc)
	if err != nil {
		return nil, 0, fmt.Errorf("route the full spec: %w", err)
	}
	target, err := url.Parse(apiURL)
	if err != nil {
		return nil, 0, err
	}

	p := &specProxy{router: router}
	forward := httputil.NewSingleHostReverseProxy(target)
	forward.ModifyResponse = p.checkResponse
	forward.Transport = &certTransport{file: filepath.Join(tempDir, "pki", "controller", "api.crt")}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, err
	}
	p.listeners = append(p.listeners, l)
	go func() {
		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sdkMode {
				p.replay(w, r, true)

				return
			}
			p.checkRequest(r)
			forward.ServeHTTP(w, r)
		}))
	}()
	if sdkMode {
		// runDirect's refusals go through the SDK too, with no check of the
		// spec, which a refusal can break.
		direct, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			p.Close()

			return nil, 0, err
		}
		p.listeners = append(p.listeners, direct)
		directPort = direct.Addr().(*net.TCPAddr).Port
		go func() {
			_ = http.Serve(direct, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.replay(w, r, false)
			}))
		}()
	}

	return p, l.Addr().(*net.TCPAddr).Port, nil
}

// checkRequest validates r against the spec, and keeps its body for the
// server and for the check of the response.
func (p *specProxy) checkRequest(
	r *http.Request,
) {
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		return
	}
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	route, params, err := p.router.FindRoute(r)
	if err != nil {
		// A request the spec does not describe, such as a path that tries
		// to leave its directory, is checked by its answer: the server must
		// not serve it.
		return
	}
	in := &openapi3filter.RequestValidationInput{
		Request:    r.Clone(context.Background()),
		PathParams: params,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	in.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := openapi3filter.ValidateRequest(context.Background(), in); err != nil {
		p.add("%s %s: the request breaks the spec: %v", r.Method, r.URL.Path, err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
}

// checkResponse validates the server's answer against the spec of the
// operation it answers.
func (p *specProxy) checkResponse(
	resp *http.Response,
) error {
	r := resp.Request
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		return nil
	}
	route, params, err := p.router.FindRoute(r)
	if err != nil {
		if resp.StatusCode < http.StatusBadRequest {
			p.add("%s %s: answered %d, and no operation in the spec describes it",
				r.Method, r.URL.Path, resp.StatusCode)
		}

		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	in := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: r, PathParams: params, Route: route,
		},
		Status:  resp.StatusCode,
		Header:  resp.Header,
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
	}
	in.SetBodyBytes(body)
	if err := openapi3filter.ValidateResponse(context.Background(), in); err != nil {
		p.add(
			"%s %s: the %d response breaks the spec: %v",
			r.Method,
			r.URL.Path,
			resp.StatusCode,
			err,
		)
	}

	return nil
}

func (p *specProxy) add(
	format string,
	args ...any,
) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.violations = append(p.violations, fmt.Sprintf(format, args...))
}

// Violations are the requests and responses that broke the spec.
func (p *specProxy) Violations() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]string(nil), p.violations...)
}
