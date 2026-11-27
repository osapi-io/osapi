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
	"math"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/telemetry/instruments"
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
//     the host, the method's key field, and any input field tagged
//     telemetry:"attr";
//   - osapi_behavior_calls_total, by behavior, method and outcome;
//   - osapi_behavior_duration_seconds, by behavior and method;
//   - osapi_behavior_changes_total, by behavior and method, for writes that
//     changed the host.
//
// An input value is recorded only when its field is the method's key, set
// with Key, or is tagged telemetry:"attr". Thus a password or a file's
// contents never reaches a trace by default. A secret is never recorded.

// Attribute keys, under osapi.behavior.
const (
	AttrBehavior = attribute.Key("osapi.behavior.name")
	AttrMethod   = attribute.Key("osapi.behavior.method")
	AttrVerb     = attribute.Key("osapi.behavior.verb")
	AttrOutcome  = attribute.Key("osapi.behavior.outcome")
	AttrChanged  = attribute.Key("osapi.behavior.changed")
	// attrInput prefixes a recorded input field: osapi.behavior.input.key.
	attrInput = "osapi.behavior.input."
)

// Outcomes a call ends in.
const (
	OutcomeOK    = "ok"
	OutcomeError = "error"
)

// SpanName is the name of the span a method's call opens on the agent.
func (b *Behavior) SpanName(
	m *Method,
) string {
	return "behavior." + b.Name + "." + m.Name
}

// SpanInputs are the input fields, by JSON name, that a call of m records on
// its span: the fields of its path first, then each field tagged telemetry:"attr". A
// secret is never one of them.
func (m *Method) SpanInputs() []string {
	t := m.In
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	var keys, tagged []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() || Secret(f) {
			continue
		}
		name := JSONName(f)
		switch {
		case m.InPath(name):
			keys = append(keys, name)
		case f.Tag.Get("telemetry") == "attr":
			tagged = append(tagged, name)
		}
	}
	return append(keys, tagged...)
}

type meters struct {
	calls    metric.Int64Counter
	duration metric.Float64Histogram
	changes  metric.Int64Counter
}

var (
	instMu sync.RWMutex
	inst   *meters
)

// SetMeterProvider creates the behavior instruments on mp. The agent calls it
// with the provider it exports metrics through. Until then the global meter
// provider is used, which records nothing unless one was installed.
func SetMeterProvider(
	mp metric.MeterProvider,
) {
	meter := mp.Meter("osapi-behavior")
	i := &meters{}
	i.calls, _ = meter.Int64Counter(
		instruments.BehaviorCalls.Name,
		metric.WithDescription(instruments.BehaviorCalls.Doc),
	)
	i.duration, _ = meter.Float64Histogram(
		instruments.BehaviorDuration.Name,
		metric.WithDescription(instruments.BehaviorDuration.Doc),
		metric.WithUnit(instruments.BehaviorDuration.Unit),
	)
	i.changes, _ = meter.Int64Counter(
		instruments.BehaviorChanges.Name,
		metric.WithDescription(instruments.BehaviorChanges.Doc),
	)

	instMu.Lock()
	inst = i
	instMu.Unlock()
}

func current() *meters {
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
//
// A write takes the lock of its behavior first, so two jobs never write one
// behavior's state at once. The agent runs several jobs at a time, and a
// write such as route.create or sshkey.add reads a file, changes it and
// writes it again: two at once would lose one change. Reads and exec take no
// lock, so a long command does not hold back other jobs. A behavior that a
// write calls through Call takes no lock of its own, so one write is never
// held behind a second write that it called.
func (m *Method) observe(
	ctx context.Context,
	provider any,
	input []byte,
) (any, error) {
	if m.Verb == types.VerbWrite {
		unlock := lockWrites(m.behaviorName())
		defer unlock()
	}

	return m.run(ctx, func(ctx context.Context) (any, any, error) {
		return m.invoke(ctx, provider, input)
	})
}

// writeLocks holds one lock for each behavior that a job writes.
var writeLocks sync.Map

// lockWrites takes the lock of the writes of a behavior, and gives the
// function that gives it back.
func lockWrites(
	behavior string,
) func() {
	mu, _ := writeLocks.LoadOrStore(behavior, &sync.Mutex{})
	l := mu.(*sync.Mutex)
	l.Lock()

	return l.Unlock
}

// behaviorName is the name of the behavior that owns the method.
func (m *Method) behaviorName() string {
	if m.owner != nil {
		return m.owner.Name
	}

	return m.Name
}

// run runs call, the method with its input, inside the span of the method,
// and records its metrics. call returns the output, and the input so that
// the span can have its tagged fields.
func (m *Method) run(
	ctx context.Context,
	call func(ctx context.Context) (out any, in any, err error),
) (any, error) {
	name := m.behaviorName()

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
		span.SetAttributes(InputAttributes(in, m.SpanInputs())...)
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

	if changed, ok := Changed(out); ok && m.Verb == types.VerbWrite {
		span.SetAttributes(AttrChanged.Bool(changed))
		if changed {
			i.changes.Add(ctx, 1, attrs)
		}
	}

	return out, err
}

// InputAttributes returns a span attribute for each field of in named in
// fields, by JSON name, keyed by that name. A secret is never recorded.
func InputAttributes(
	in any,
	fields []string,
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
		// dsl behavior validate refuses that pairing.
		name := JSONName(f)
		if !f.IsExported() || Secret(f) || !slices.Contains(fields, name) {
			continue
		}
		key := attribute.Key(attrInput + name)
		switch fv := v.Field(i); fv.Kind() {
		case reflect.String:
			attrs = append(attrs, key.String(fv.String()))
		case reflect.Bool:
			attrs = append(attrs, key.Bool(fv.Bool()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			attrs = append(attrs, key.Int64(fv.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			attrs = append(attrs, uintAttr(key, fv.Uint()))
		case reflect.Float32, reflect.Float64:
			attrs = append(attrs, key.Float64(fv.Float()))
		case reflect.Slice, reflect.Array:
			if fv.Len() == 0 {
				continue
			}
			if a, ok := sliceAttr(key, fv); ok {
				attrs = append(attrs, a)
			}
		}
	}

	return attrs
}

// uintAttr is u as an int attribute, or as text when it is too large for
// one.
func uintAttr(
	key attribute.Key,
	u uint64,
) attribute.KeyValue {
	if u > math.MaxInt64 {
		return key.String(strconv.FormatUint(u, 10))
	}

	return key.Int64(int64(u))
}

// sliceAttr is a list of strings, bools, ints or floats as a list
// attribute, such as the servers of ntp. ok is false for a list of
// anything else, such as structs.
func sliceAttr(
	key attribute.Key,
	fv reflect.Value,
) (attribute.KeyValue, bool) {
	n := fv.Len()
	switch fv.Type().Elem().Kind() {
	case reflect.String:
		out := make([]string, n)
		for i := range n {
			out[i] = fv.Index(i).String()
		}

		return key.StringSlice(out), true
	case reflect.Bool:
		out := make([]bool, n)
		for i := range n {
			out[i] = fv.Index(i).Bool()
		}

		return key.BoolSlice(out), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out := make([]int64, n)
		for i := range n {
			out[i] = fv.Index(i).Int()
		}

		return key.Int64Slice(out), true
	case reflect.Float32, reflect.Float64:
		out := make([]float64, n)
		for i := range n {
			out[i] = fv.Index(i).Float()
		}

		return key.Float64Slice(out), true
	default:
		return attribute.KeyValue{}, false
	}
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
		sudo, runs := x.Programs("")
		for _, p := range append(append(sudo, runs...), x.RunsIfPresent...) {
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
