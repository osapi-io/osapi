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
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TagBefore is the struct tag of an input field that the controller step of
// its method sets, before:"set". A caller does not send the field: it is not
// in the request, the command or the UI, and the step fills it from the
// fields that the caller sends.
const TagBefore = "before"

// SetByBefore reports whether the controller step of f's method sets f.
func SetByBefore(
	f reflect.StructField,
) bool {
	_, ok := f.Tag.Lookup(TagBefore)

	return ok
}

// Before gives the method a controller step: fn runs in the generated route
// on the controller, before osapi queues the job, as
// func(ctx context.Context, in In) (In, error). It changes the input that
// the caller sent into the input that the agent gets, such as a password
// into its hash. A secret.Value field that the step leaves set goes to the
// job sealed to the target agents, when the controller has PKI. dsl
// behavior validate checks fn's signature.
func (m *Method) Before(
	fn any,
) *Method {
	m.before = reflect.ValueOf(fn)

	return m
}

// HasBefore reports whether the method has a controller step.
func (m *Method) HasBefore() bool {
	return m.before.IsValid()
}

// RunBefore is in after the controller step of b's method m, in a span of
// its own on the controller. Without a step, it is in. It fails when the
// step fails.
func (b *Behavior) RunBefore(
	ctx context.Context,
	m *Method,
	in any,
) (any, error) {
	if !m.HasBefore() {
		return in, nil
	}
	ctx, span := otel.Tracer("osapi-behavior").Start(
		ctx,
		b.SpanName(m)+".before",
		trace.WithAttributes(AttrBehavior.String(b.Name), AttrMethod.String(m.Name)),
	)
	defer span.End()

	out := m.before.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(in)})
	if err, _ := out[1].Interface().(error); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, fmt.Errorf("the controller step of %s %s: %w", b.Name, m.Name, err)
	}
	result := out[0]

	return result.Interface(), nil
}

// beforeProblems are the RuleBefore problems of m: a step that is not
// func(context.Context, In) (In, error), a step on a method that the API
// does not serve, a step with no secret field to keep out of the job, and a
// field tagged before that has no step to set it, or a value other than set.
func beforeProblems(
	b *Behavior,
	m *Method,
) []string {
	var out []string
	var set []string
	if m.In.Kind() == reflect.Struct {
		for i := range m.In.NumField() {
			f := m.In.Field(i)
			if v, ok := f.Tag.Lookup(TagBefore); ok {
				set = append(set, JSONName(f))
				if v != "set" {
					out = append(
						out,
						fmt.Sprintf(
							"input field %s is tagged before:%q; the only value is \"set\"",
							JSONName(f),
							v,
						),
					)
				}
			}
		}
	}
	if !m.HasBefore() {
		for _, name := range set {
			out = append(
				out,
				"input field "+name+" is tagged before:\"set\", but the method has no .Before step to set it",
			)
		}

		return out
	}
	want := reflect.FuncOf(
		[]reflect.Type{reflect.TypeFor[context.Context](), m.In},
		[]reflect.Type{m.In, reflect.TypeFor[error]()},
		false,
	)
	if m.before.Kind() != reflect.Func || m.before.Type() != want {
		out = append(
			out,
			fmt.Sprintf("the .Before step is %v; it must be %v", m.before.Type(), want),
		)
	}
	if !b.API {
		out = append(
			out,
			"the method has a .Before step, but the behavior is not served: the step runs in its route",
		)
	}
	hasSecret := false
	if m.In.Kind() == reflect.Struct {
		for i := range m.In.NumField() {
			if Secret(m.In.Field(i)) {
				hasSecret = true
			}
		}
	}
	if !hasSecret {
		out = append(
			out,
			"the method has a .Before step, but no input field is a secret.Value for it to keep out of the job",
		)
	}

	return out
}
