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

package endpoint

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// ExamplesExtension is the extension of an operation of the spec that has
// the examples of its method: for each, the doc and the input, by the name
// of each parameter and property of the body. The SDK's generator writes a
// program from them that calls the method with each input.
const ExamplesExtension = "x-osapi-examples"

// SpecExample is one example of a method in the spec.
type SpecExample struct {
	// Doc says what the example shows.
	Doc string `json:"doc"`
	// Input is the value of each parameter and each property of the body
	// that the example's flags set, by its name in the spec.
	Input map[string]any `json:"input"`
}

// bytesType is the type of a file's content, such as the file of an upload.
// Its flag names the path of the file, a string.
var (
	bytesType  = reflect.TypeFor[[]byte]()
	stringType = reflect.TypeFor[string]()
)

// exampleInput is where a flag of a method's command puts its value.
type exampleInput struct {
	// name is the name of the parameter or the property in the spec.
	name string
	// typ is the field's type.
	typ reflect.Type
}

// specExamples are the examples of m as the spec has them: the flags of
// each one read into the input that the command sends. Two examples that
// send the same input, such as one with --json, are one. An example with no
// doc has the summary, and the flags after the first one, so that the docs
// of two examples differ. A file of a multipart form, such as the upload of
// a file, is the path that its flag names.
func specExamples(
	m *Method,
) ([]SpecExample, error) {
	inputs := exampleInputs(m.In)
	var out []SpecExample
	seen := map[string]bool{}
	for _, e := range m.examples {
		in, err := readFlags(inputs, e.Flags)
		if err != nil {
			return nil, fmt.Errorf("example %q: %w", e.Flags, err)
		}
		key := fmt.Sprint(in)
		if seen[key] {
			continue
		}
		seen[key] = true
		doc := e.Doc
		if doc == "" {
			doc = m.summary
			if len(out) > 0 {
				doc += ", with " + strings.TrimSpace(e.Flags)
			}
		}
		out = append(out, SpecExample{Doc: doc, Input: in})
	}

	return out, nil
}

// exampleInputs are the flags of the command of a method whose input is in,
// as command.InputFlags names them, each with where its value goes.
func exampleInputs(
	in reflect.Type,
) map[string]exampleInput {
	out := map[string]exampleInput{}
	add := func(f reflect.StructField) {
		name := command.FlagName(f)
		if tag := f.Tag.Get("flag"); tag != "" {
			name = tag
		}
		v := exampleInput{name: jsonName(f), typ: f.Type}
		out[name] = v
		if short := f.Tag.Get("short"); short != "" {
			out[short] = v
		}
	}
	for _, f := range fields(in) {
		if f.Tag.Get("in") == "body" && f.Type.Kind() == reflect.Struct {
			for _, bf := range fields(f.Type) {
				add(bf)
			}

			continue
		}
		add(f)
	}

	return out
}

// readFlags is the input that a command line of flags sets, by the name of
// each parameter and property.
func readFlags(
	inputs map[string]exampleInput,
	line string,
) (map[string]any, error) {
	words, err := splitWords(line)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for i := 0; i < len(words); i++ {
		word := words[i]
		name, ok := strings.CutPrefix(word, "-")
		if !ok || name == "" {
			return nil, fmt.Errorf("%s is not a flag", word)
		}
		name = strings.TrimPrefix(name, "-")
		name, value, hasValue := strings.Cut(name, "=")
		in, ok := inputs[name]
		if !ok {
			// A flag of each command, such as --json, sets nothing of the
			// input. Validate reports a flag that the command does not
			// have, so the spec skips it, with its value.
			if !command.GlobalFlag(name) && !hasValue && i+1 < len(words) &&
				!strings.HasPrefix(words[i+1], "-") {
				i++
			}

			continue
		}
		t := in.typ
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t == bytesType {
			t = stringType
		}
		if !hasValue {
			if t.Kind() == reflect.Bool {
				out[in.name] = true

				continue
			}
			if i+1 == len(words) {
				return nil, fmt.Errorf("%s has no value", word)
			}
			i++
			value = words[i]
		}
		if err := setValue(out, in.name, t, value); err != nil {
			return nil, fmt.Errorf("%s: %w", word, err)
		}
	}

	return out, nil
}

// setValue puts text, the value of a flag, in out under name, as the type t
// holds it. A list adds it to the values before it, and a map takes it as
// key=value.
func setValue(
	out map[string]any,
	name string,
	t reflect.Type,
	text string,
) error {
	switch t.Kind() {
	case reflect.Slice:
		v, err := scalar(t.Elem(), text)
		if err != nil {
			return err
		}
		list, _ := out[name].([]any)
		out[name] = append(list, v)
	case reflect.Map:
		key, value, ok := strings.Cut(text, "=")
		if !ok {
			return fmt.Errorf("%s is not key=value", text)
		}
		m, _ := out[name].(map[string]any)
		if m == nil {
			m = map[string]any{}
		}
		m[key] = value
		out[name] = m
	default:
		v, err := scalar(t, text)
		if err != nil {
			return err
		}
		out[name] = v
	}

	return nil
}

// scalar is text as a value of t: a number for a number, a bool for a bool,
// and the text for anything else.
func scalar(
	t reflect.Type,
	text string,
) (any, error) {
	switch t.Kind() {
	case reflect.Bool:
		return strconv.ParseBool(text)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseInt(text, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(text, 64)
	default:
		return text, nil
	}
}

// splitWords is the words of a command line: a word in double quotes is
// one word, as a Go string literal, as the docs write a value with a space.
func splitWords(
	line string,
) ([]string, error) {
	var out []string
	rest := strings.TrimSpace(line)
	for rest != "" {
		if rest[0] != '"' {
			word, after, _ := strings.Cut(rest, " ")
			out = append(out, word)
			rest = strings.TrimSpace(after)

			continue
		}
		end := closingQuote(rest)
		if end < 0 {
			return nil, fmt.Errorf("%s has no closing quote", rest)
		}
		word, err := strconv.Unquote(rest[:end+1])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rest[:end+1], err)
		}
		out = append(out, word)
		rest = strings.TrimSpace(rest[end+1:])
	}

	return out, nil
}

// closingQuote is the index of the quote that ends the quoted word at the
// start of s, or -1.
func closingQuote(
	s string,
) int {
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}

	return -1
}
