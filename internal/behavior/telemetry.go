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

package behavior

import (
	"context"
	"reflect"
	"slices"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Every method call gets the same telemetry, written here once so a behavior
// writes none of it:
//
//   - a span, behavior.<name>.<method>, under the agent's job.process span,
//     with the behavior, method, verb and outcome, whether a write changed
//     the host, and any input field tagged telemetry:"attr";
//   - osapi_behavior_calls_total, by behavior, method and outcome;
//   - osapi_behavior_duration_seconds, by behavior and method;
//   - osapi_behavior_changes_total, by behavior and method, for writes that
//     changed the host.
//
// An input value is recorded only when its field is tagged telemetry:"attr",
// so a password or a file's contents never reaches a trace by default.

// Attribute keys, under osapi.behavior.
const (
	AttrBehavior = attribute.Key("osapi.behavior.name")
	AttrMethod   = attribute.Key("osapi.behavior.method")
	AttrVerb     = attribute.Key("osapi.behavior.verb")
	AttrOutcome  = attribute.Key("osapi.behavior.outcome")
	AttrChanged  = attribute.Key("osapi.behavior.changed")
	// attrInput prefixes a tagged input field: osapi.behavior.input.key.
	attrInput = "osapi.behavior.input."
)

// Outcomes a call ends in.
const (
	OutcomeOK    = "ok"
	OutcomeError = "error"
)

// The metrics every behavior method records, and the span every call opens.
const (
	MetricCalls    = "osapi_behavior_calls_total"
	MetricDuration = "osapi_behavior_duration_seconds"
	MetricChanges  = "osapi_behavior_changes_total"
)

// SpanName is the name of the span a method's call opens on the agent.
func (b *Behavior) SpanName(
	m *Method,
) string {
	return "behavior." + b.Name + "." + m.Name
}

// TaggedInputs are the input fields of t, by JSON name, tagged
// telemetry:"attr": the ones a call records on its span.
func TaggedInputs(
	t reflect.Type,
) []string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var names []string
	for i := range t.NumField() {
		f := t.Field(i)
		if f.IsExported() && f.Tag.Get("telemetry") == "attr" && !Secret(f) {
			names = append(names, JSONName(f))
		}
	}

	return names
}

type instruments struct {
	calls    metric.Int64Counter
	duration metric.Float64Histogram
	changes  metric.Int64Counter
}

var (
	instMu sync.RWMutex
	inst   *instruments
)

// SetMeterProvider creates the behavior instruments on mp. The agent calls it
// with the provider it exports metrics through. Until then the global meter
// provider is used, which records nothing unless one was installed.
func SetMeterProvider(
	mp metric.MeterProvider,
) {
	meter := mp.Meter("osapi-behavior")
	i := &instruments{}
	i.calls, _ = meter.Int64Counter(
		MetricCalls,
		metric.WithDescription("Behavior method calls, by behavior, method and outcome"),
	)
	i.duration, _ = meter.Float64Histogram(
		MetricDuration,
		metric.WithDescription("Behavior method duration in seconds"),
		metric.WithUnit("s"),
	)
	i.changes, _ = meter.Int64Counter(
		MetricChanges,
		metric.WithDescription("Behavior writes that changed the host"),
	)

	instMu.Lock()
	inst = i
	instMu.Unlock()
}

func current() *instruments {
	instMu.RLock()
	i := inst
	instMu.RUnlock()
	if i != nil {
		return i
	}

	SetMeterProvider(otel.GetMeterProvider())

	return current()
}

// observe decodes input and runs the method inside its span with its
// metrics, as the agent calls it.
func (m *Method) observe(
	ctx context.Context,
	provider any,
	input []byte,
) (any, error) {
	return m.run(ctx, func(ctx context.Context) (any, any, error) {
		return m.invoke(ctx, provider, input)
	})
}

// run runs call, the method with its input, inside the span of the method,
// and records its metrics. call returns the output, and the input so that
// the span can have its tagged fields.
func (m *Method) run(
	ctx context.Context,
	call func(ctx context.Context) (out any, in any, err error),
) (any, error) {
	name := m.Name
	if m.owner != nil {
		name = m.owner.Name
	}

	base := make([]attribute.KeyValue, 0, 4)
	base = append(base,
		AttrBehavior.String(name),
		AttrMethod.String(m.Name),
		AttrVerb.String(string(m.Verb)),
	)

	ctx, span := otel.Tracer("osapi-behavior").Start(
		ctx,
		"behavior."+name+"."+m.Name, // SpanName, without needing the owner
		trace.WithAttributes(base...),
	)
	defer span.End()

	start := time.Now()
	out, in, err := call(ctx)
	elapsed := time.Since(start).Seconds()

	if in != nil {
		span.SetAttributes(InputAttributes(in)...)
	}

	outcome := OutcomeOK
	if err != nil {
		outcome = OutcomeError
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.SetAttributes(AttrOutcome.String(outcome))

	i := current()
	attrs := metric.WithAttributes(base...)
	i.calls.Add(ctx, 1, metric.WithAttributes(append(base, AttrOutcome.String(outcome))...))
	i.duration.Record(ctx, elapsed, attrs)

	if changed, ok := Changed(out); ok && m.Verb == VerbWrite {
		span.SetAttributes(AttrChanged.Bool(changed))
		if changed {
			i.changes.Add(ctx, 1, attrs)
		}
	}

	return out, err
}

// InputAttributes returns a span attribute for each field of in tagged
// telemetry:"attr", keyed by its JSON name.
func InputAttributes(
	in any,
) []attribute.KeyValue {
	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	var attrs []attribute.KeyValue
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		// A secret is never recorded, whatever its telemetry tag says;
		// behavior validate refuses that pairing.
		if !f.IsExported() || f.Tag.Get("telemetry") != "attr" || Secret(f) {
			continue
		}
		key := attribute.Key(attrInput + JSONName(f))
		switch fv := v.Field(i); fv.Kind() {
		case reflect.String:
			attrs = append(attrs, key.String(fv.String()))
		case reflect.Bool:
			attrs = append(attrs, key.Bool(fv.Bool()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			attrs = append(attrs, key.Int64(fv.Int()))
		case reflect.Float32, reflect.Float64:
			attrs = append(attrs, key.Float64(fv.Float()))
		}
	}

	return attrs
}

// Changed reports a write's changed field: the bool field whose JSON name is
// changed. ok is false when out has none.
func Changed(
	out any,
) (changed bool, ok bool) {
	v := reflect.ValueOf(out)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false, false
	}
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if JSONName(f) == "changed" && f.Type.Kind() == reflect.Bool {
			return v.Field(i).Bool(), true
		}
	}

	return false, false
}

// ChildSpans are the spans a call of b can make under its own span, from what
// b and the behaviors it requires declare: the span of each required
// behavior's methods, a command span for each program, the file spans when a
// behavior declares Writes, and the kv spans of the file-state bucket when
// one declares KeepsState. Each appears once, in that order.
func ChildSpans(
	b *Behavior,
) []string {
	var out []string
	add := func(s string) {
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}

	var walk func(*Behavior, bool)
	walk = func(x *Behavior, required bool) {
		if required {
			add("behavior." + x.Name + ".<method>")
		}
		for _, r := range x.Requires {
			walk(r, true)
		}
	}
	walk(b, false)

	all := func(x *Behavior, f func(*Behavior)) {
		var each func(*Behavior)
		each = func(y *Behavior) {
			f(y)
			for _, r := range y.Requires {
				each(r)
			}
		}
		each(x)
	}
	all(b, func(x *Behavior) {
		for _, p := range append(append([]string{}, x.Sudo...), x.Runs...) {
			add("exec " + p)
		}
	})
	all(b, func(x *Behavior) {
		if len(x.Writes) > 0 {
			add("file.write")
			add("file.remove")
			add("file.chmod")
		}
	})
	all(b, func(x *Behavior) {
		if x.KeepsState {
			add("kv.get")
			add("kv.put")
		}
	})

	return out
}
