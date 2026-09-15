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
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	cronparser "github.com/robfig/cron/v3"

	"github.com/osapi-io/osapi/internal/facts"
)

var instance = validator.New()

// sysctlKeyPattern matches dot-separated sysctl parameter names: letters,
// digits, dots, underscores, and hyphens. Slashes, spaces, and other
// characters are rejected so the key cannot be used for path traversal.
// The first character must be a letter or digit so the key cannot be
// mistaken for a `sysctl` command-line option (e.g. -p) when passed as
// an argument.
var sysctlKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

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

	// sysctl_key validates a sysctl parameter key: dot-separated names only,
	// which rejects path traversal and other injection characters when the
	// key is used to build a file path.
	_ = instance.RegisterValidation("sysctl_key", func(fl validator.FieldLevel) bool {
		return sysctlKeyPattern.MatchString(fl.Field().String())
	})

	// no_linebreak rejects values containing a carriage return or line feed,
	// which would let the value inject additional lines into a config file.
	_ = instance.RegisterValidation("no_linebreak", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), "\r\n")
	})
}

// customHints maps validator tags to a hint appended to the default error.
var customHints = map[string]func(fe validator.FieldError) string{
	"valid_target": func(fe validator.FieldError) string {
		if t, ok := IsPendingTarget(); ok {
			if t == "_all" || t == "_any" {
				return "all agents are pending PKI enrollment — accept them with: osapi client agent accept --hostname <hostname>"
			}

			return fmt.Sprintf(
				"agent %q is pending PKI enrollment — accept it with: osapi client agent accept --hostname %s",
				t,
				t,
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
	"sysctl_key": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"%q is not a valid sysctl key (expected: %s)",
			fe.Value(),
			sysctlKeyPattern.String(),
		)
	},
	"no_linebreak": func(_ validator.FieldError) string {
		return "value must not contain line breaks"
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
	if rv.Kind() == reflect.Pointer {
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
		case reflect.Pointer, reflect.Slice, reflect.Map:
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
