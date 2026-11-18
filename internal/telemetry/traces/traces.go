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

// Package traces reads the spans of a trace from a tracing backend, so that
// osapi can show a trace without the caller reaching the backend.
package traces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"go.opentelemetry.io/otel/attribute"
)

// ErrNotFound is returned when the backend has no trace with the ID. The
// backend gets the spans a few seconds after the request, so a trace of a
// request that just ran can be not found yet.
var ErrNotFound = errors.New("trace not found")

// Trace is the spans of one trace, ordered by start time.
type Trace struct {
	ID    string
	Spans []Span
}

// Span is one span of a trace.
type Span struct {
	ID            string
	ParentID      string
	Name          string
	Service       string
	Kind          string
	Start         time.Time
	Duration      time.Duration
	Failed        bool
	StatusMessage string
	Attributes    map[string]string
}

// Jaeger reads traces from the query API of Jaeger, /api/v3/traces/{id},
// which answers OTLP JSON.
type Jaeger struct {
	base   string
	client *http.Client
}

// NewJaeger returns a reader for the Jaeger at base, such as
// http://jaeger:16686, the address of its UI and its query API.
func NewJaeger(
	base string,
	client *http.Client,
) *Jaeger {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Jaeger{base: strings.TrimSuffix(base, "/"), client: client}
}

// read returns the trace id from Jaeger, or ErrNotFound.
func (j *Jaeger) read(
	ctx context.Context,
	id string,
) (*Trace, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		j.base+"/api/v3/traces/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("read trace %s: %w", id, err)
	}
	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("read trace %s: %w", id, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read trace %s: %w", id, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read trace %s: jaeger answered %d", id, resp.StatusCode)
	}

	var doc otlpDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("read trace %s: %w", id, err)
	}

	t := &Trace{ID: id}
	for _, rs := range doc.Result.ResourceSpans {
		service := attributes(rs.Resource.Attributes)["service.name"]
		for _, ss := range rs.ScopeSpans {
			for _, s := range ss.Spans {
				t.Spans = append(t.Spans, s.span(service))
			}
		}
	}
	if len(t.Spans) == 0 {
		return nil, ErrNotFound
	}
	sort.SliceStable(t.Spans, func(a, b int) bool {
		return t.Spans[a].Start.Before(t.Spans[b].Start)
	})

	return t, nil
}

// otlpDoc is the part of the OTLP JSON answer of Jaeger that osapi reads.
type otlpDoc struct {
	Result struct {
		ResourceSpans []struct {
			Resource struct {
				Attributes []otlpAttr `json:"attributes"`
			} `json:"resource"`
			ScopeSpans []struct {
				Spans []otlpSpan `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	} `json:"result"`
}

type otlpSpan struct {
	SpanID       string     `json:"spanId"`
	ParentSpanID string     `json:"parentSpanId"`
	Name         string     `json:"name"`
	Kind         any        `json:"kind"`
	Start        string     `json:"startTimeUnixNano"`
	End          string     `json:"endTimeUnixNano"`
	Attributes   []otlpAttr `json:"attributes"`
	Status       struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

type otlpAttr struct {
	Key   string         `json:"key"`
	Value map[string]any `json:"value"`
}

// span is s as a Span of service.
func (s otlpSpan) span(
	service string,
) Span {
	start := nanos(s.Start)
	code := strings.ToUpper(fmt.Sprint(s.Status.Code))

	return Span{
		ID:            s.SpanID,
		ParentID:      s.ParentSpanID,
		Name:          s.Name,
		Service:       service,
		Kind:          kind(s.Kind),
		Start:         start,
		Duration:      nanos(s.End).Sub(start),
		Failed:        code == "STATUS_CODE_ERROR" || code == "2",
		StatusMessage: s.Status.Message,
		Attributes:    attributes(s.Attributes),
	}
}

// nanos is the time of a count of nanoseconds since the epoch, which OTLP
// JSON gives as text.
func nanos(
	v string,
) time.Time {
	n, _ := strconv.ParseInt(v, 10, 64)

	return time.Unix(0, n).UTC()
}

// kind is the span kind as a word, such as server, from the name or the
// number OTLP JSON gives.
func kind(
	v any,
) string {
	names := []string{"", "internal", "server", "client", "producer", "consumer"}
	switch k := v.(type) {
	case float64:
		if int(k) > 0 && int(k) < len(names) {
			return names[int(k)]
		}
	case string:
		return strings.ToLower(strings.TrimPrefix(k, "SPAN_KIND_"))
	}

	return ""
}

// attributes are attrs as text, one value for each key.
func attributes(
	attrs []otlpAttr,
) map[string]string {
	out := make(map[string]string, len(attrs))
	for _, a := range attrs {
		for _, v := range a.Value {
			out[a.Key] = text(v)
		}
	}

	return out
}

// text is an OTLP JSON value as text. An array is its values, joined.
func text(
	v any,
) string {
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		values, _ := t["values"].([]any)
		parts := make([]string, 0, len(values))
		for _, e := range values {
			m, _ := e.(map[string]any)
			for _, x := range m {
				parts = append(parts, text(x))
			}
		}

		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(t)
	}
}

// Trace returns the trace id from Jaeger, or ErrNotFound, in a traces.read
// span: the read is an HTTP call to Jaeger, which no other span shows.
func (j *Jaeger) Trace(
	ctx context.Context,
	id string,
) (*Trace, error) {
	ctx, span := opspan.Start(ctx, "osapi-traces", "traces.read",
		attribute.String("osapi.trace.id", id),
		attribute.String("server.address", j.base),
	)
	t, err := j.read(ctx, id)
	if err == nil {
		span.SetAttributes(attribute.Int("osapi.trace.spans", len(t.Spans)))
	}

	return t, opspan.End(span, err)
}
