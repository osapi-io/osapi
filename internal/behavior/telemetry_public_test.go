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

// secretKey has a key that is a secret, beside a tagged field.
type secretKey struct {
	Key  string `json:"key"  secret:"true"`
	Name string `json:"name"               telemetry:"attr"`
}

func readTagged(context.Context, fakeProvider, tagged) (entry, error) {
	return entry{}, nil
}

func readSecretKey(context.Context, fakeProvider, secretKey) (entry, error) {
	return entry{}, nil
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

func (s *TelemetryPublicTestSuite) TestCall() {
	b := top()

	s.Run("runs the method inside its span, under the caller", func() {
		s.SetupTest()
		ctx, parent := otel.Tracer("test").Start(context.Background(), "behavior.other.create")

		out, err := behavior.Call(ctx, b, get, fakeProvider{prefix: "v-"}, getIn{Key: "k"})
		parent.End()

		s.Require().NoError(err)
		s.Equal(entry{Key: "k", Value: "v-k"}, out)
		spans := s.spans.Ended()
		s.Require().Len(spans, 2)
		s.Equal("behavior.sysctl.list", spans[0].Name(), "the first method declared with get")
		s.Equal(parent.SpanContext().SpanID(), spans[0].Parent().SpanID())
		s.EqualValues(1, s.metrics()[behavior.MetricCalls])
	})

	s.Run("a failed call marks its span", func() {
		s.SetupTest()

		_, err := behavior.Call(context.Background(), b, get, fakeProvider{}, getIn{Key: "boom"})

		s.EqualError(err, "boom")
		s.Equal(codes.Error, s.spans.Ended()[0].Status().Code)
	})

	s.Run("an input that breaks a validate rule is an error, and the method does not run", func() {
		s.SetupTest()

		_, err := behavior.Call(context.Background(), b, get, fakeProvider{}, getIn{})

		s.ErrorContains(err, "method list")
		s.ErrorContains(err, "required")
		s.Equal(codes.Error, s.spans.Ended()[0].Status().Code, "the refusal is on the span")
	})

	s.Run(
		"Check refuses an input that breaks a rule, or does not decode, and passes the rest",
		func() {
			m := b.Method("get")

			s.ErrorContains(m.Check(context.Background(), []byte(`{"key":""}`)), "required")
			s.ErrorContains(m.Check(context.Background(), []byte(`{`)), "decode input")
			s.NoError(m.Check(context.Background(), []byte(`{"key":"k"}`)))
			s.ErrorContains(
				m.Check(context.Background(), nil),
				"required",
				"no input is the zero input",
			)

			bare := behavior.Read("status", readBare)
			s.NoError(
				bare.Check(context.Background(), []byte(`"x"`)),
				"an input that is not a struct has no tags to check",
			)
		},
	)

	s.Run("a function that is not a method of the behavior is an error", func() {
		s.SetupTest()

		_, err := behavior.Call(context.Background(), grouped(), set, fakeProvider{}, getIn{})

		s.ErrorContains(err, "is not a method of behavior dns")
		s.Empty(s.spans.Ended())
	})
}

func (s *TelemetryPublicTestSuite) TestSetMeterProvider() {
	s.Run("when no provider was set the global one is used", func() {
		behavior.ResetInstruments()
		_, err := top().Method("get").Invoke(context.Background(), fakeProvider{}, nil)
		s.NoError(err)
	})
}

func (s *TelemetryPublicTestSuite) TestInputAttributes() {
	all := []string{"key", "on", "count", "ratio", "list", "token"}

	tests := []struct {
		name   string
		in     any
		fields []string
		want   map[string]any
	}{
		{
			name: "when fields are named each kind is recorded and a secret is not",
			in: tagged{
				Key: "k", On: true, Count: 3, Ratio: 0.5, Secret: "hunter2", Token: "s3cret",
			},
			fields: all,
			want: map[string]any{
				"osapi.behavior.input.key":   "k",
				"osapi.behavior.input.on":    true,
				"osapi.behavior.input.count": int64(3),
				"osapi.behavior.input.ratio": 0.5,
			},
		},
		{
			name:   "when the input is a pointer it is followed",
			in:     &tagged{Key: "k"},
			fields: all,
			want: map[string]any{
				"osapi.behavior.input.key":   "k",
				"osapi.behavior.input.on":    false,
				"osapi.behavior.input.count": int64(0),
				"osapi.behavior.input.ratio": float64(0),
			},
		},
		{
			name:   "when a field is not named it is not recorded",
			in:     tagged{Key: "k", On: true},
			fields: []string{"key"},
			want:   map[string]any{"osapi.behavior.input.key": "k"},
		},
		{
			name: "when the pointer is nil",
			in:   (*tagged)(nil), fields: all, want: map[string]any{},
		},
		{name: "when the input is not a struct", in: "x", fields: all, want: map[string]any{}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := map[string]any{}
			for _, kv := range behavior.InputAttributes(tc.in, tc.fields) {
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

func (s *TelemetryPublicTestSuite) TestSpanInputs() {
	tests := []struct {
		name string
		m    *behavior.Method
		want []string
	}{
		{
			name: "when the method has a key it comes first, then the tagged fields",
			m:    behavior.Read("get", readTagged).Key("on"),
			want: []string{"on", "key", "count", "ratio", "list"},
		},
		{
			name: "when the key is not tagged it is still recorded",
			m:    behavior.Read("get", get).Key("key"),
			want: []string{"key"},
		},
		{
			name: "when the method has no key only the tagged fields are recorded",
			m:    behavior.Read("get", readTagged),
			want: []string{"key", "on", "count", "ratio", "list"},
		},
		{
			name: "when nothing is tagged and there is no key",
			m:    behavior.Read("get", get),
			want: nil,
		},
		{
			name: "when the key is a secret it is not recorded",
			m:    behavior.Read("get", readSecretKey).Key("key"),
			want: []string{"name"},
		},
		{
			name: "when the input is a pointer it is followed",
			m:    &behavior.Method{In: reflect.TypeFor[*getIn](), KeyField: "key"},
			want: []string{"key"},
		},
		{
			name: "when the input is not a struct",
			m:    &behavior.Method{In: reflect.TypeFor[string]()},
			want: nil,
		},
		{name: "when there is no input", m: &behavior.Method{}, want: nil},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.m.SpanInputs())
		})
	}
}

func (s *TelemetryPublicTestSuite) TestSpanRecordsKey() {
	s.Run("when a method has a key its span records it without a tag", func() {
		s.SetupTest()

		_, err := top().Method("get").
			Invoke(context.Background(), fakeProvider{}, []byte(`{"key":"vm.swappiness"}`))
		s.Require().NoError(err)

		spans := s.spans.Ended()
		s.Require().Len(spans, 1)
		got := map[string]any{}
		for _, kv := range spans[0].Attributes() {
			got[string(kv.Key)] = kv.Value.AsInterface()
		}
		s.Equal("vm.swappiness", got["osapi.behavior.input.key"])
	})
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
