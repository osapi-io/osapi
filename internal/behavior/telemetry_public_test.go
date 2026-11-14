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

package behavior_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/osapi-io/osapi/internal/behavior"
)

type TelemetryPublicTestSuite struct {
	suite.Suite

	spans  *tracetest.SpanRecorder
	reader *sdkmetric.ManualReader
}

func (s *TelemetryPublicTestSuite) SetupTest() {
	s.spans = tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.spans)))
	s.reader = sdkmetric.NewManualReader()
	behavior.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(s.reader)))
}

type tagged struct {
	Key    string  `json:"key"    telemetry:"attr"`
	On     bool    `json:"on"     telemetry:"attr"`
	Count  int     `json:"count"  telemetry:"attr"`
	Ratio  float64 `json:"ratio"  telemetry:"attr"`
	List   []int   `json:"list"   telemetry:"attr"`
	Secret string  `json:"secret"`
	Token  string  `json:"token"  telemetry:"attr" secret:"true"`
}

func (s *TelemetryPublicTestSuite) metrics() map[string]int64 {
	var rm metricdata.ResourceMetrics
	s.Require().NoError(s.reader.Collect(context.Background(), &rm))

	out := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch d := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, p := range d.DataPoints {
					out[m.Name] += p.Value
				}
			case metricdata.Histogram[float64]:
				for _, p := range d.DataPoints {
					out[m.Name] += int64(p.Count)
				}
			}
		}
	}

	return out
}

func (s *TelemetryPublicTestSuite) TestInvoke() {
	b := top()
	tests := []struct {
		name         string
		method       string
		input        string
		validateFunc func(spans []sdktrace.ReadOnlySpan, metrics map[string]int64)
	}{
		{
			name:   "when a write changes the host",
			method: "create",
			input:  `{"key":"vm.swappiness"}`,
			validateFunc: func(spans []sdktrace.ReadOnlySpan, metrics map[string]int64) {
				s.Require().Len(spans, 1)
				s.Equal("behavior.sysctl.create", spans[0].Name())
				attrs := attribute.NewSet(spans[0].Attributes()...)
				v, _ := attrs.Value(behavior.AttrChanged)
				s.True(v.AsBool())
				v, _ = attrs.Value(behavior.AttrOutcome)
				s.Equal(behavior.OutcomeOK, v.AsString())
				s.EqualValues(1, metrics[behavior.MetricCalls])
				s.EqualValues(1, metrics[behavior.MetricDuration])
				s.EqualValues(1, metrics[behavior.MetricChanges])
			},
		},
		{
			name:   "when a read fails",
			method: "get",
			input:  `{"key":"boom"}`,
			validateFunc: func(spans []sdktrace.ReadOnlySpan, metrics map[string]int64) {
				s.Require().Len(spans, 1)
				s.Equal(codes.Error, spans[0].Status().Code)
				attrs := attribute.NewSet(spans[0].Attributes()...)
				_, ok := attrs.Value(behavior.AttrChanged)
				s.False(ok, "a read records no changed")
				s.EqualValues(1, metrics[behavior.MetricCalls])
				s.Zero(metrics[behavior.MetricChanges])
			},
		},
		{
			name:   "when the input does not decode",
			method: "get",
			input:  `{"key":`,
			validateFunc: func(spans []sdktrace.ReadOnlySpan, _ map[string]int64) {
				s.Require().Len(spans, 1)
				s.Equal(codes.Error, spans[0].Status().Code)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			_, _ = b.Method(tc.method).
				Invoke(context.Background(), fakeProvider{}, json.RawMessage(tc.input))
			tc.validateFunc(s.spans.Ended(), s.metrics())
		})
	}
}

func (s *TelemetryPublicTestSuite) TestSetMeterProvider() {
	s.Run("when no provider was set the global one is used", func() {
		behavior.ResetInstruments()
		_, err := top().Method("get").Invoke(context.Background(), fakeProvider{}, nil)
		s.NoError(err)
	})
}

func (s *TelemetryPublicTestSuite) TestInputAttributes() {
	tests := []struct {
		name string
		in   any
		want map[string]any
	}{
		{
			name: "when fields are tagged each kind is recorded and untagged ones are not",
			in: tagged{
				Key: "k", On: true, Count: 3, Ratio: 0.5, Secret: "hunter2", Token: "s3cret",
			},
			want: map[string]any{
				"osapi.behavior.input.key":   "k",
				"osapi.behavior.input.on":    true,
				"osapi.behavior.input.count": int64(3),
				"osapi.behavior.input.ratio": 0.5,
			},
		},
		{
			name: "when the input is a pointer it is followed",
			in:   &tagged{Key: "k"},
			want: map[string]any{
				"osapi.behavior.input.key":   "k",
				"osapi.behavior.input.on":    false,
				"osapi.behavior.input.count": int64(0),
				"osapi.behavior.input.ratio": float64(0),
			},
		},
		{name: "when the pointer is nil", in: (*tagged)(nil), want: map[string]any{}},
		{name: "when the input is not a struct", in: "x", want: map[string]any{}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := map[string]any{}
			for _, kv := range behavior.InputAttributes(tc.in) {
				got[string(kv.Key)] = kv.Value.AsInterface()
			}
			s.Equal(tc.want, got)
		})
	}
}

func (s *TelemetryPublicTestSuite) TestChanged() {
	tests := []struct {
		name        string
		out         any
		wantChanged bool
		wantOK      bool
	}{
		{
			name:        "when the output has changed",
			out:         result{Changed: true},
			wantChanged: true,
			wantOK:      true,
		},
		{name: "when the output has no changed field", out: entry{}, wantOK: false},
		{name: "when the output is not a struct", out: []entry{}, wantOK: false},
		{name: "when there is no output", out: nil, wantOK: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			changed, ok := behavior.Changed(tc.out)
			s.Equal(tc.wantChanged, changed)
			s.Equal(tc.wantOK, ok)
		})
	}
}

func (s *TelemetryPublicTestSuite) TestTaggedInputs() {
	tests := []struct {
		name string
		t    reflect.Type
		want []string
	}{
		{
			name: "when fields are tagged",
			t:    reflect.TypeFor[tagged](),
			want: []string{"key", "on", "count", "ratio", "list"},
		},
		{
			name: "when given a pointer",
			t:    reflect.TypeFor[*tagged](),
			want: []string{"key", "on", "count", "ratio", "list"},
		},
		{name: "when nothing is tagged", t: reflect.TypeFor[getIn](), want: nil},
		{name: "when not a struct", t: reflect.TypeFor[string](), want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, behavior.TaggedInputs(tc.t))
		})
	}
}

func (s *TelemetryPublicTestSuite) TestSpanName() {
	b := top()
	s.Equal("behavior.sysctl.get", b.SpanName(b.Method("get")))
}

func TestTelemetryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TelemetryPublicTestSuite))
}
