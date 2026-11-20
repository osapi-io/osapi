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

package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/validation"
)

// UsageError is a command that was called wrong, such as a flag whose
// value breaks a rule of its field. The CLI prints it with a pointer to the
// help, and exits 2.
type UsageError struct {
	// Msg says what is wrong, and names the flag.
	Msg string
}

func (e *UsageError) Error() string {
	return e.Msg
}

// Run runs method of g with in, the input that the command's flags made:
// it checks in against its validate tags, calls the method's handler with
// the deps that Provide gave, and gives the output to Output, which prints
// it. An input that breaks a rule is a *UsageError, and the handler does not
// run. The generated command of each local method calls it.
func Run(
	cmd *cobra.Command,
	g *LocalGroup,
	method string,
	in any,
) error {
	m := g.Method(method)
	if m == nil {
		return fmt.Errorf("the %s group has no method %s", g.Name, method)
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := Check(ctx, in); err != nil {
		return err
	}
	d, err := g.deps()
	if err != nil {
		return err
	}

	results := reflect.ValueOf(m.handler).Call([]reflect.Value{
		reflect.ValueOf(ctx), d, reflect.ValueOf(in),
	})
	if err, _ := results[1].Interface().(error); err != nil {
		return err
	}
	body, err := json.Marshal(results[0].Interface())
	if err != nil {
		return fmt.Errorf("encode %s %s: %w", g.Name, method, err)
	}
	Output(g.Name, method, in, body)

	return nil
}

// deps is the value of the group's Deps: the one Provide gives, or the zero
// value of a Deps with no field. A Deps with fields that nothing gives is a
// mistake in the wiring.
func (g *LocalGroup) deps() (reflect.Value, error) {
	switch {
	case g.provide != nil:
		return reflect.ValueOf(g.provide()), nil
	case g.Deps == nil:
		return reflect.Value{}, fmt.Errorf("the %s group Uses no type", g.Name)
	case g.Deps.Kind() == reflect.Struct && g.Deps.NumField() > 0:
		return reflect.Value{}, fmt.Errorf(
			"the %s group's %v is not given: add command.Provide to cmd/commands.go",
			g.Name,
			g.Deps,
		)
	}

	return reflect.New(g.Deps).Elem(), nil
}

// Check validates in against the validate tags of its fields, as Run does
// before the handler runs. A value that breaks a rule is a *UsageError that
// names the flag of its field.
func Check(
	ctx context.Context,
	in any,
) error {
	err := validation.Instance().StructCtx(ctx, in)
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err
	}
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		msgs = append(msgs, problem(ctx, reflect.TypeOf(in), fe))
	}

	return &UsageError{Msg: strings.Join(msgs, "; ")}
}

// problem says which flag breaks which rule, and how.
func problem(
	ctx context.Context,
	t reflect.Type,
	fe validator.FieldError,
) string {
	_, path, _ := strings.Cut(fe.StructNamespace(), ".")
	field, _, _ := strings.Cut(path, "[")
	flag := "--" + field
	if f, ok := t.FieldByName(field); ok {
		flag = "--" + FlagName(f)
	}
	rule := fe.Tag()
	if fe.Param() != "" {
		rule += "=" + fe.Param()
	}

	switch {
	case fe.Tag() == "required":
		return flag + " is required"
	case validation.Hint(ctx, fe) != "":
		return flag + ": " + validation.Hint(ctx, fe)
	case fe.Tag() == "oneof":
		return fmt.Sprintf("%s: %q is not one of %s",
			flag, fmt.Sprint(fe.Value()), strings.Join(strings.Fields(fe.Param()), ", "))
	default:
		return fmt.Sprintf("%s: %q breaks %s", flag, fmt.Sprint(fe.Value()), rule)
	}
}

// FlagName is the flag of an input field: its json name, with _ as -.
func FlagName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		name = f.Name
	}

	return strings.ReplaceAll(name, "_", "-")
}
