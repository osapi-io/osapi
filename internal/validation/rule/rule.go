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

// Package rule is what a validate tag means in a spec. The endpoint DSL and
// the behavior DSL both read a field's validate tag through it, so each one
// puts the same constraints in its spec, and each refuses a rule that the
// spec cannot show.
package rule

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// Extension is the OpenAPI extension that carries a field's whole validate
// tag. oapi-codegen copies it onto the generated field, so a Go client made
// from the spec validates as the server does.
const Extension = "x-oapi-codegen-extra-tags"

// Target is a schema that takes the constraints of a validate tag.
type Target interface {
	// Enum sets the values the field may have.
	Enum(values []string)
	// Pattern sets the regular expression the value matches.
	Pattern(re string)
	// Format sets the JSON Schema format of the value.
	Format(format string)
	// Bound sets min, max or len: the length of a string, the items of an
	// array, or the value of a number.
	Bound(rule string, n float64)
}

var (
	mu      sync.RWMutex
	schemas = map[string]func(t Target){}
)

// Register declares what a custom validate tag means in a spec. A custom
// validator is a Go function, and nothing can read a schema out of it, so its
// schema is registered beside it:
//
//	rule.Register("sysctl_key", rule.Pattern(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`))
//
// A rule that the spec cannot show more of than its type registers
// rule.None. Apply refuses a tag that is neither standard nor registered.
func Register(
	tag string,
	apply func(t Target),
) {
	mu.Lock()
	defer mu.Unlock()

	schemas[tag] = apply
}

// Pattern is a schema that sets a regular expression.
func Pattern(
	re string,
) func(t Target) {
	return func(t Target) { t.Pattern(re) }
}

// Format is a schema that sets a JSON Schema format.
func Format(
	format string,
) func(t Target) {
	return func(t Target) { t.Format(format) }
}

// None is the schema of a rule that the spec cannot show more of than the
// field's type, such as a target that the controller resolves.
func None(Target) {}

// flags change how validation runs, not what a value may be.
var flags = map[string]bool{
	"omitempty": true,
	"required":  true,
	"dive":      true,
}

// formats maps the standard validate tags to JSON Schema formats.
var formats = map[string]string{
	"ip":       "ip",
	"ipv4":     "ipv4",
	"ipv6":     "ipv6",
	"url":      "uri",
	"uri":      "uri",
	"email":    "email",
	"hostname": "hostname",
	"uuid":     "uuid",
	"cidr":     "cidr",
}

// Apply sets the constraints of a validate tag on t, and reports whether the
// field is required. A rule with no schema is an error, so the spec never
// checks less than the server does.
func Apply(
	tag string,
	t Target,
) (bool, error) {
	if tag == "" {
		return false, nil
	}

	for _, r := range strings.Split(tag, ",") {
		name, param, _ := strings.Cut(r, "=")
		switch {
		case flags[name]:
		case name == "min" || name == "max" || name == "len":
			n, err := strconv.ParseFloat(param, 64)
			if err != nil {
				return false, fmt.Errorf("validate %s=%s: not a number", name, param)
			}
			t.Bound(name, n)
		case name == "oneof":
			t.Enum(strings.Fields(param))
		case name == "startswith":
			t.Pattern("^" + regexp.QuoteMeta(param))
		case formats[name] != "":
			t.Format(formats[name])
		default:
			mu.RLock()
			apply, ok := schemas[name]
			mu.RUnlock()
			if !ok {
				return false, fmt.Errorf(
					"validate tag %q has no schema; register one with rule.Register", name)
			}
			apply(t)
		}
	}

	return Has(tag, "required"), nil
}

// Has reports whether the validate tag has the rule.
func Has(
	tag string,
	name string,
) bool {
	return slices.Contains(strings.Split(tag, ","), name)
}

// Requires and Fix are the check that each DSL runs on its inputs, in the
// same words, so endpoints and behaviors hold the same rule.
const (
	Requires = "Each input field has a validate tag: its rules, or omitempty for an optional field with none. " +
		"The spec carries the whole tag, and the server answers 400 to a value that breaks it."
	Fix = `Add validate:"..." to the field, such as validate:"required,min=1,max=255", ` +
		`or validate:"omitempty" for an optional field with no other rule.`
)

// Untagged is the name of each input field of t that has no validate tag: an
// exported field that JSON reads, and each field of a struct that is the
// whole body (in:"body").
func Untagged(
	t reflect.Type,
) []string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var out []string
	for i := range t.NumField() {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if !f.IsExported() || name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		if _, ok := f.Tag.Lookup("validate"); !ok {
			out = append(out, name)
		}
		if f.Tag.Get("in") == "body" {
			for _, inner := range Untagged(f.Type) {
				out = append(out, name+"."+inner)
			}
		}
	}

	return out
}

// The words of the validate tag and of Register in the spec of each DSL,
// osapi controller endpoint spec and osapi behavior spec, so an author of
// either reads the same rule.
const (
	TagName = `validate:"..."`
	TagDoes = "The rules the server checks before the method runs: a value that breaks one is a 400 " +
		`with {"error": "..."}. required makes the field required, oneof gives an enum, min, max and ` +
		"len the length of a string, the items of an array or the bounds of a number, startswith a " +
		"pattern, and uuid, email, hostname, ip, ipv4, ipv6, cidr, url and uri a format. The spec " +
		"carries these, and the whole tag in x-oapi-codegen-extra-tags. A custom rule, such as " +
		"sysctl_key, needs its schema from rule.Register; a rule with none fails the spec."
	TagWhen = `On each input field. A field with no limit is validate:"omitempty"; validate refuses ` +
		"an input field with no tag."

	FailsRequires = "Each rule of an input field has a .Fails input that breaks it, which the API " +
		"refuses with 400. An endpoint's required needs one, with the field empty; a behavior's " +
		"needs none, as its generated tests leave the field out. A path field's required, and a min " +
		"of 1 on a string that is required, optional or in the path, need none: their only bad " +
		"value is empty."
	FailsWhen = "For each rule of each input field, as validate requires, and for each input " +
		"that the method refuses for another reason."
	FailsFix = `Add .Fails("<what is wrong>", <Input>{...}) to the method, with the other fields valid ` +
		`and the field's value breaking the rule that the message names, such as a name of 256 ` +
		`characters for max=255.`

	RegisterName = `rule.Register("tag", rule.Pattern(re))`
	RegisterDoes = "Declares what a custom validate rule means in the spec: rule.Pattern(re), " +
		"rule.Format(f), or rule.None for a rule that the spec cannot show, such as a target the " +
		"controller resolves."
	RegisterWhen = "Beside the validator of a new custom rule, in internal/validation."
)

// Custom is a validate tag that an author declares: the check the server
// runs, what the tag means in the spec, and what a caller is told when a
// value fails it.
type Custom struct {
	// Tag is the name used in validate tags, such as sysctl_key.
	Tag string
	// Check reports whether a field's value, as a string, passes.
	Check func(value string) bool
	// Schema is what the tag means in the spec.
	Schema func(t Target)
	// Message follows the failing value in the error, such as "is not a
	// valid sysctl key".
	Message string
}

var (
	customs  []Custom
	onCustom func(c Custom)
)

// Declare registers a custom rule: its schema here, and its check with the
// validator through OnDeclare.
func Declare(
	c Custom,
) {
	Register(c.Tag, c.Schema)

	mu.Lock()
	customs = append(customs, c)
	hook := onCustom
	mu.Unlock()

	if hook != nil {
		hook(c)
	}
}

// OnDeclare has fn called for every custom rule: those declared already,
// then each one after. The validation package calls it once, to register
// each rule with the validator the server uses, so a rule can be declared
// before or after that package starts.
func OnDeclare(
	fn func(c Custom),
) {
	mu.Lock()
	onCustom = fn
	declared := append([]Custom(nil), customs...)
	mu.Unlock()

	for _, c := range declared {
		fn(c)
	}
}

// Need is a rule of an input field that a method must show it refuses, with
// an input that breaks it.
type Need struct {
	// Field is the Go path of the field below the input, such as "Name" or
	// "Body.Target".
	Field string
	// Rule is the rule as the tag has it, such as "max=255".
	Rule string
}

// Key is the need as validation.Breaks reports a broken rule.
func (n Need) Key() string {
	name, _, _ := strings.Cut(n.Rule, "=")

	return n.Field + ":" + name
}

// Needs is each rule of t's input fields that needs a bad input: each rule
// but omitempty and dive. A min of 1 or less on a string that is optional,
// in the path or required is not one: its only bad value is empty, which the
// tag lets through, the route cannot carry, or required refuses first. A
// path field's required is not one either: the route cannot carry an empty
// key. Required is one only when withRequired: the behavior DSL's generated
// tests leave each required field out, and the endpoint DSL has no such
// tests.
func Needs(
	t reflect.Type,
	withRequired bool,
) []Need {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var out []Need
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() || strings.HasPrefix(f.Tag.Get("json"), "-") {
			continue
		}
		tag := f.Tag.Get("validate")
		for _, r := range strings.Split(tag, ",") {
			name, param, _ := strings.Cut(r, "=")
			switch {
			case r == "", name == "omitempty", name == "dive",
				emptyOnly(f, tag, name, param):
				continue
			case name == "required" && (!withRequired || f.Tag.Get("in") == "path"):
				continue
			}
			out = append(out, Need{Field: f.Name, Rule: r})
		}
		if f.Tag.Get("in") == "body" {
			for _, n := range Needs(f.Type, withRequired) {
				n.Field = f.Name + "." + n.Field
				out = append(out, n)
			}
		}
	}

	return out
}

// emptyOnly reports whether the rule's only bad value is an empty string
// that cannot reach it: a min of 1 or less on a string that is optional, in
// the path, or required, where an empty value breaks required first.
func emptyOnly(
	f reflect.StructField,
	tag string,
	name string,
	param string,
) bool {
	if name != "min" || f.Type.Kind() != reflect.String {
		return false
	}
	n, err := strconv.ParseFloat(param, 64)

	return err == nil && n <= 1 &&
		(Has(tag, "omitempty") || Has(tag, "required") || f.Tag.Get("in") == "path")
}
