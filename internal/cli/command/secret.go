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
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/secret"
)

// Stdin is the value of a secret flag that reads its value from standard
// input.
const Stdin = "-"

// SecretHelp is what the help of a secret flag adds: the forms it takes.
const SecretHelp = " (a secret: secret://<vault>/<key>, or - to read it from standard input)"

// SecretsHelp is what the help of a flag of a map of secrets adds.
const SecretsHelp = " (secrets: NAME=secret://<vault>/<key>)"

// errLiteral is a literal value in a secret flag: the shell history keeps
// it, and ps shows it to each user of the host.
var errLiteral = errors.New(
	"takes a reference, secret://<vault>/<key>, or - to read the value from " +
		"standard input; a value in a flag stays in the shell history")

// Secret is the value of the secret flag name, as the command sends it:
// a reference as it is, or the value that standard input gives for -, with
// no line break at its end. A literal value is refused, and so is a value
// from standard input that is empty. A refusal is reported, and ok is
// false.
func Secret(
	cmd *cobra.Command,
	name string,
	flag string,
) (secret.Value, bool) {
	v := secret.Value(flag)
	switch {
	case flag == "" || v.IsReference():
		return v, true
	case flag != Stdin:
		deps.Error(fmt.Errorf("--%s %w", name, errLiteral))

		return "", false
	}
	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		deps.Error(fmt.Errorf("--%s: read standard input: %w", name, err))

		return "", false
	}
	value := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	if value == "" {
		deps.Error(fmt.Errorf("--%s -: standard input is empty", name))

		return "", false
	}

	return secret.Value(value), true
}

// Secrets are the values of the flag name of a map of secrets, each
// NAME=secret://<vault>/<key>. A literal value is refused, as for Secret,
// and standard input gives one value only, so - is refused too. A refusal
// is reported, and ok is false.
func Secrets(
	_ *cobra.Command,
	name string,
	flags map[string]string,
) (map[string]secret.Value, bool) {
	if flags == nil {
		return nil, true
	}
	out := make(map[string]secret.Value, len(flags))
	for _, k := range slices.Sorted(maps.Keys(flags)) {
		v := secret.Value(flags[k])
		if !v.IsReference() {
			deps.Error(fmt.Errorf("--%s %s takes a reference, secret://<vault>/<key>; "+
				"a value in a flag stays in the shell history", name, k))

			return nil, false
		}
		out[k] = v
	}

	return out, true
}
