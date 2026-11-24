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

// Package secret has the type of a secret in the three DSLs: Value, a
// reference, secret://<vault>/<key>, or a literal value. A field of this
// type is a secret: the spec makes it write-only, validate keeps it out of
// a span and a URL, the CLI refuses a literal in a flag, and each record
// shows the reference, or [redacted] for a literal. Only Reveal gives the
// value.
package secret

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// Scheme starts each reference: secret://<vault>/<key>.
const Scheme = "secret://"

// Redacted is what a literal value shows as, in each record.
const Redacted = "[redacted]"

// Takes is what a secret field takes, as its description in the spec, the
// docs and method get says.
const Takes = "A secret: a reference, secret://<vault>/<key>, or the value."

// Describe is doc, the description of a secret field, with Takes.
func Describe(
	doc string,
) string {
	if doc == "" {
		return Takes
	}

	return strings.TrimSuffix(doc, ".") + ". " + Takes
}

// Value is a secret: a reference, such as secret://default/registry-token,
// or a literal value. String, Format, MarshalJSON and LogValue give the
// reference, or Redacted for a literal, so a log line, a span, a chain
// record or an audit entry never has the value. Reveal gives it, where a
// provider gives it to a command's standard input, an environment or a
// file. The route of each method resolves a reference from its vault,
// before the handler gets the input.
//
// It is a string to JSON and to the validate tags, so a request carries it
// as a string and no_linebreak checks it as one. A conversion to string
// gives the value too: use Reveal, which a reader can find.
type Value string

// Reveal is the value: the literal, or the reference as it is.
func (v Value) Reveal() string {
	return string(v)
}

// IsReference reports whether v is a reference, secret://<vault>/<key>,
// and not a literal value.
func (v Value) IsReference() bool {
	return strings.HasPrefix(string(v), Scheme)
}

// IsZero reports whether v is empty.
func (v Value) IsZero() bool {
	return v == ""
}

// String is the reference, Redacted for a literal, or "" when v is empty.
func (v Value) String() string {
	switch {
	case v == "":
		return ""
	case v.IsReference():
		return string(v)
	default:
		return Redacted
	}
}

// Format prints String with each verb, so %v, %s, %q and %#v never print a
// literal value.
func (v Value) Format(
	f fmt.State,
	verb rune,
) {
	_, _ = fmt.Fprintf(f, fmt.FormatString(f, verb), v.String())
}

// MarshalJSON is String as a JSON string. Plain gives a copy of an input
// whose secrets marshal their values, for the request and the job.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// Unresolved is the error of the reference ref, which no vault resolves.
// A command of the command DSL has no vault yet, so it refuses each
// reference with it: a reference is never sent as the value.
func Unresolved(
	ref string,
) error {
	vault, _, _ := strings.Cut(strings.TrimPrefix(ref, Scheme), "/")

	return fmt.Errorf("%s: the vault %q does not exist", ref, vault)
}

// Parse is the vault and the key of the reference ref,
// secret://<vault>/<key>. A ref that is not a reference, or that has no
// vault or no key, is an error that gives the form.
func Parse(
	ref string,
) (string, string, error) {
	vault, key, ok := strings.Cut(strings.TrimPrefix(ref, Scheme), "/")
	if !strings.HasPrefix(ref, Scheme) || !ok || vault == "" || key == "" {
		return "", "", fmt.Errorf("%s: a reference is %s<vault>/<key>", ref, Scheme)
	}

	return vault, key, nil
}

// LogValue is String, so slog never logs a literal value.
func (v Value) LogValue() slog.Value {
	return slog.StringValue(v.String())
}
