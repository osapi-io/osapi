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

package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/validation"
)

// The rules that Load checks, by the name that config validate reports.
const (
	RuleFile    = "file"
	RuleParse   = "parse"
	RuleValue   = "value"
	RuleSetting = "setting"
	RuleBucket  = "bucket"
	RuleVault   = "vault"
)

// Checks are the rules that Load applies to a config file, in the order that
// it applies them. config validate --help prints each one, and the problem
// that config validate reports carries the fix of its rule. Each Rule
// constant has one entry; a test fails on a rule without one.
var Checks = guide.Checks{
	{
		Rule:     RuleFile,
		Requires: "The config file exists, and osapi can read it.",
		Fix:      "Give the path with -f, or put the file at /etc/osapi/osapi.yaml. Check that the user that runs osapi can read it.",
	},
	{
		Rule:     RuleParse,
		Requires: "The file is YAML, and each value has the type of its key.",
		Fix:      "Correct the YAML at the line that the message names. osapi config list shows the type of each key.",
	},
	{
		Rule:     RuleValue,
		Requires: "Each value passes the validate rule of its key, with the defaults and the environment applied.",
		Fix:      "Set the key to a value that its rule allows. osapi config get --key <key> shows its rule, type and default.",
	},
	{
		Rule:     RuleSetting,
		Requires: "Each setting of a behavior, under behaviors.<name>, passes its validate rule.",
		Fix:      "Set the key to a value that its rule allows. osapi dsl behavior get --name <name> lists the settings of the behavior.",
	},
	{
		Rule:     RuleBucket,
		Requires: "Each NATS bucket has its own name. A shared bucket mixes the keys of two roles.",
		Fix:      "Give one of the two keys that the message names a different bucket.",
	},
	{
		Rule:     RuleVault,
		Requires: "Each vault of controller.secrets.vaults has its own name and its own key_file, a type of vault that osapi knows, a config block that its type takes, and what its type needs, such as the key of a nats vault.",
		Fix:      "Correct the vault that the message names. osapi config get --key controller.secrets.vaults describes a vault, and the message names the types and the settings that there are.",
	},
}

// Error is a failure of Load, with the rule that it breaks.
type Error struct {
	Rule string
	Err  error
}

// Error is the message of the failure.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Unwrap is the failure without its rule.
func (e *Error) Unwrap() error {
	return e.Err
}

// RuleOf is the rule that err breaks, or "" when err is not an Error.
func RuleOf(
	err error,
) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Rule
	}

	return ""
}

// broke is err with rule, or nil when err is nil.
func broke(
	rule string,
	err error,
) error {
	if err == nil {
		return nil
	}

	return &Error{Rule: rule, Err: err}
}

// Problem is one rule that a config breaks.
type Problem struct {
	Rule string `json:"rule"`
	// Key is the config key that breaks the rule, such as
	// controller.api.port, when the rule is about one key.
	Key     string `json:"key,omitempty"`
	Message string `json:"message"`
	// Fix is how to fix it, from the check of Rule.
	Fix string `json:"fix" table:"-"`
}

// String is the problem as one line: its key, then its message.
func (p Problem) String() string {
	if p.Key == "" {
		return p.Message
	}

	return p.Key + " " + p.Message
}

// joined is problems as one error, each problem on its own after a "; ".
func joined(
	problems []Problem,
) error {
	msgs := make([]string, 0, len(problems))
	for _, p := range problems {
		msgs = append(msgs, p.String())
	}

	return errors.New(strings.Join(msgs, "; "))
}

// withFixes sets the fix of each problem from its rule.
func withFixes(
	problems []Problem,
) []Problem {
	for i := range problems {
		problems[i].Fix = Checks.Fix(problems[i].Rule)
	}

	return problems
}

// valueProblems are the problems of err, from the validator, under rule: one
// for each field, named by its config key under prefix, such as
// controller.api.port, not by its Go field, with the validate rule that it
// breaks. t is the type that was checked. Thus the key is what osapi config
// get --key takes. An error that is not from the validator is one problem.
func valueProblems(
	err error,
	rule string,
	t reflect.Type,
	prefix string,
) []Problem {
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return []Problem{{Rule: rule, Message: err.Error()}}
	}

	problems := make([]Problem, 0, len(errs))
	for _, fe := range errs {
		tag := fe.ActualTag()
		// An "or" rule such as startswith=a|startswith=b holds its parameters.
		if fe.Param() != "" && !strings.Contains(tag, "=") {
			tag += "=" + fe.Param()
		}
		msg := fmt.Sprintf("breaks its rule %q", tag)
		if hint := validation.Hint(context.Background(), fe); hint != "" {
			msg += ": " + hint
		}
		problems = append(problems, Problem{
			Rule:    rule,
			Key:     keyOf(t, fe.Namespace(), prefix),
			Message: msg,
		})
	}

	return problems
}

// keyOf is the config key of namespace, a field path from the validator such
// as Config.Controller.API.Port, under prefix. Each Go field becomes its
// mapstructure name, and an index such as [0] stays. A field that t does not
// have keeps its Go name.
func keyOf(
	t reflect.Type,
	namespace string,
	prefix string,
) string {
	parts := strings.Split(namespace, ".")[1:]
	names := make([]string, 0, len(parts)+1)
	if prefix != "" {
		names = append(names, prefix)
	}
	for _, part := range parts {
		field, index, _ := strings.Cut(part, "[")
		if index != "" {
			index = "[" + index
		}
		for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Map) {
			t = t.Elem()
		}
		name := field
		if t != nil && t.Kind() == reflect.Struct {
			if f, ok := t.FieldByName(field); ok {
				if tag, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ","); tag != "" {
					name = tag
				}
				t = f.Type
			} else {
				t = nil
			}
		}
		names = append(names, name+index)
	}

	return strings.Join(names, ".")
}
