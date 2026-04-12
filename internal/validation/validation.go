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

// Package validation provides a shared validator instance.
package validation

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	cronparser "github.com/robfig/cron/v3"

	"github.com/retr0h/osapi/internal/facts"
)

var instance = validator.New()

func init() {
	// alphanum_or_fact accepts alphanumeric values or @fact. prefixed references
	// with a known fact key. Fact references are resolved agent-side.
	_ = instance.RegisterValidation("alphanum_or_fact", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if strings.HasPrefix(v, facts.Prefix) {
			return facts.IsKnownKey(v[len(facts.Prefix):])
		}
		return instance.Var(v, "alphanum") == nil
	})

	// ip_or_fact accepts IP addresses (v4/v6) or @fact. prefixed references
	// with a known fact key. Fact references are resolved agent-side.
	_ = instance.RegisterValidation("ip_or_fact", func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		if strings.HasPrefix(v, facts.Prefix) {
			return facts.IsKnownKey(v[len(facts.Prefix):])
		}
		return instance.Var(v, "ip") == nil
	})

	// go_duration validates a Go time.Duration string (e.g., "30s", "5m", "1h", "720h").
	_ = instance.RegisterValidation("go_duration", func(fl validator.FieldLevel) bool {
		_, err := time.ParseDuration(fl.Field().String())
		return err == nil
	})

	// cron_schedule validates a standard 5-field cron expression
	// (minute hour day-of-month month day-of-week).
	cronParser := cronparser.NewParser(
		cronparser.Minute | cronparser.Hour | cronparser.Dom | cronparser.Month | cronparser.Dow,
	)
	_ = instance.RegisterValidation("cron_schedule", func(fl validator.FieldLevel) bool {
		_, err := cronParser.Parse(fl.Field().String())
		return err == nil
	})
}

// customHints maps validator tags to a hint appended to the default error.
var customHints = map[string]func(fe validator.FieldError) string{
	"valid_target": func(fe validator.FieldError) string {
		if t, ok := IsPendingTarget(); ok {
			return fmt.Sprintf(
				"agent %q is pending PKI enrollment — accept it with: osapi client agent accept --hostname %s",
				t, t,
			)
		}

		return fmt.Sprintf("target agent %q not found", fe.Value())
	},
	"go_duration": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid Go duration (e.g., 30s, 5m, 1h)",
			fe.Value(),
		)
	},
	"cron_schedule": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid cron expression (expected: minute hour day-of-month month day-of-week)",
			fe.Value(),
		)
	},
}

// Struct validates a struct and returns the error message and false if invalid.
func Struct(
	v any,
) (string, bool) {
	if err := instance.Struct(v); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return formatErrors(validationErrors), false
	}

	return "", true
}

// Var validates a single variable against a tag string and returns the error
// message and false if invalid. This is useful for path parameters that
// oapi-codegen does not generate validate tags for in strict-server mode.
func Var(
	field any,
	tag string,
) (string, bool) {
	if err := instance.Var(field, tag); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return formatErrors(validationErrors), false
	}

	return "", true
}

// formatErrors builds the error string, appending a custom hint for known
// tags while keeping the standard validator prefix.
func formatErrors(
	errs validator.ValidationErrors,
) string {
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		msg := fe.Error()
		if fn, ok := customHints[fe.Tag()]; ok {
			msg = fmt.Sprintf("%s: %s", msg, fn(fe))
		}
		msgs = append(msgs, msg)
	}

	return strings.Join(msgs, "; ")
}

// AtLeastOneField checks that at least one exported pointer, slice, or map
// field in the struct is non-nil, or that at least one non-pointer field is
// non-zero. Returns an error message and false if all fields are nil/zero
// (i.e., the update body is empty).
func AtLeastOneField(
	v any,
) (string, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return "expected struct", false
	}

	rt := rv.Type()
	for i := range rt.NumField() {
		field := rv.Field(i)
		if !rt.Field(i).IsExported() {
			continue
		}

		switch field.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Map:
			if !field.IsNil() {
				return "", true
			}
		default:
			if !field.IsZero() {
				return "", true
			}
		}
	}

	return "at least one field must be provided", false
}

// Instance returns the shared validator for registering custom validators.
func Instance() *validator.Validate {
	return instance
}
