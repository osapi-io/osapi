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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// sdkMode is OSAPI_INTEGRATION_CLIENT=sdk: the suite runs each case, refusal,
// check and scenario as it does with the CLI, and the proxy sends each
// request that the CLI makes through the SDK's typed method for its
// operation (osapi-io/osapi#1599). The CLI only reads the flags and prints
// the answer. The request that reaches the controller is the one that the
// SDK builds, and the answer that the checks read is the one that the SDK
// decoded. The spec checks the SDK's request and the controller's answer.
var sdkMode = os.Getenv("OSAPI_INTEGRATION_CLIENT") == "sdk"

// sdkUsed is the number of requests that went through each typed method,
// which the end of the run reports.
var sdkUsed = struct {
	sync.Mutex
	calls map[string]int
}{calls: map[string]int{}}

// reportSDK prints how many requests went through how many typed methods,
// and is 1 when none did: the suite then ran with no request through the
// SDK.
func reportSDK() int {
	if !sdkMode {
		return 0
	}
	sdkUsed.Lock()
	defer sdkUsed.Unlock()
	total := 0
	for _, n := range sdkUsed.calls {
		total += n
	}
	fmt.Fprintf(os.Stderr, "integration: %d requests went through %d typed methods of the SDK\n",
		total, len(sdkUsed.calls))
	if total == 0 {
		return 1
	}

	return 0
}

// directPort is the port of the proxy without the spec, which runDirect
// calls in the SDK mode: a refusal can break the spec.
var directPort int

func init() {
	// The CLI sends its trace in traceparent. The SDK sends the trace of its
	// context, so the proxy reads the CLI's into it.
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

// checkedTransport checks each request of the SDK and its answer against
// the spec, as the proxy checks the CLI's in the CLI mode.
type checkedTransport struct {
	p    *specProxy
	next http.RoundTripper
}

func (t checkedTransport) RoundTrip(
	r *http.Request,
) (*http.Response, error) {
	if r.Body == nil {
		r.Body = http.NoBody
	}
	t.p.checkRequest(r)
	resp, err := t.next.RoundTrip(r)
	if err == nil {
		err = t.p.checkResponse(resp)
	}

	return resp, err
}

// replay answers r, a request of the CLI, with the answer of the SDK's
// typed method for its operation. A request that no operation of the spec
// describes goes through the SDK's Send: no typed method can send it. With
// checked, the spec checks the SDK's request and its answer.
func (p *specProxy) replay(
	w http.ResponseWriter,
	r *http.Request,
	checked bool,
) {
	var base http.RoundTripper = &certTransport{
		file: filepath.Join(tempDir, "pki", "controller", "api.crt"),
	}
	if checked {
		base = checkedTransport{p: p, next: base}
	}
	bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	c := client.New(controllerURL, bearer, client.WithHTTPTransport(base))
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	body, _ := io.ReadAll(r.Body)

	route, params, err := p.router.FindRoute(r)
	if err != nil {
		resp, err := c.Send(ctx, r.Method, r.URL.RequestURI(), r.Header.Get("Content-Type"), body)
		writeSDKAnswer(w, resp, err, http.StatusOK)

		return
	}
	op := r.Method + " " + route.Path
	calls, _ := sdkCalls()
	k, ok := calls[op]
	if !ok {
		p.add("%s %s: the SDK has no typed method for %s", r.Method, r.URL.Path, op)
		http.Error(w, "no typed method of the SDK for "+op, http.StatusNotImplemented)

		return
	}
	fn, err := k.fn(c)
	if err != nil {
		p.add("%s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, err.Error(), http.StatusNotImplemented)

		return
	}
	req := sdkRequest{path: params, query: r.URL.Query(), body: body}
	if k.method.Multipart {
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			http.Error(w, "the form: "+err.Error(), http.StatusBadRequest)

			return
		}
		req.form = r.MultipartForm
	}
	sdkUsed.Lock()
	sdkUsed.calls[k.String()]++
	sdkUsed.Unlock()
	args, err := k.args(fn.Type(), req)
	if err != nil {
		// The CLI sent what the SDK cannot: the SDK suite fails, and the CLI
		// sees a refusal.
		p.add("%s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
	out := fn.Call(append([]reflect.Value{reflect.ValueOf(ctx)}, args...))
	status := http.StatusOK
	for _, s := range k.method.Statuses {
		if s >= 200 && s < 300 {
			status = s

			break
		}
	}
	errOut, _ := out[len(out)-1].Interface().(error)
	if len(out) == 1 {
		writeSDKAnswer(w, nil, errOut, status)

		return
	}
	if errOut != nil {
		writeSDKAnswer(w, nil, errOut, status)

		return
	}
	resp := out[0]
	raw, _ := resp.MethodByName("RawJSON").Call(nil)[0].Interface().([]byte)
	if raw == nil {
		// A Response that the SDK made itself, such as Upload's answer for a
		// file that the Object Store has: its data is the answer.
		raw, _ = json.Marshal(resp.Elem().FieldByName("Data").Interface())
	}
	trace, _ := resp.MethodByName("TraceID").Call(nil)[0].Interface().(string)
	writeSDKAnswer(w, client.NewResponse(json.RawMessage(raw), raw), nil, status, trace)
}

// writeSDKAnswer writes the answer of the SDK to the CLI: the body of resp,
// or the status and the body of err, an error of the API. Another error,
// such as an answer that does not decode into the method's type, is a 502,
// which the CLI reports.
func writeSDKAnswer(
	w http.ResponseWriter,
	resp *client.Response[json.RawMessage],
	err error,
	status int,
	trace ...string,
) {
	if len(trace) > 0 && trace[0] != "" {
		w.Header().Set("traceresponse", "00-"+trace[0]+"-0000000000000001-01")
	} else if resp != nil && resp.TraceID() != "" {
		w.Header().Set("traceresponse", "00-"+resp.TraceID()+"-0000000000000001-01")
	}
	if err != nil {
		var api *client.APIError
		if !errors.As(err, &api) {
			proxy.add("the SDK: %v", err)
			http.Error(w, fmt.Sprintf("the SDK: %v", err), http.StatusBadGateway)

			return
		}
		code, body := api.Answer()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write(body)

		return
	}
	if resp == nil || len(resp.RawJSON()) == 0 {
		if status == http.StatusOK {
			status = http.StatusNoContent
		}
		w.WriteHeader(status)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(resp.RawJSON())
}

// clientName is the client that the suite sends its requests through.
func clientName() string {
	if sdkMode {
		return "sdk"
	}

	return "cli"
}
