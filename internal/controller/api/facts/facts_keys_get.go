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

package facts

import (
	"context"

	"github.com/retr0h/osapi/internal/controller/api/facts/gen"
	factskeys "github.com/retr0h/osapi/internal/facts"
)

// builtInDescriptions provides human-readable descriptions for built-in keys.
var builtInDescriptions = map[string]string{
	factskeys.KeyInterfacePrimary: "Primary network interface name",
	factskeys.KeyHostname:         "Agent hostname",
	factskeys.KeyArch:             "CPU architecture",
	factskeys.KeyKernel:           "Kernel version",
	factskeys.KeyFQDN:             "Fully qualified domain name",
}

// GetFactKeys returns the list of available fact keys.
func (f *Facts) GetFactKeys(
	_ context.Context,
	_ gen.GetFactKeysRequestObject,
) (gen.GetFactKeysResponseObject, error) {
	builtInKeys := factskeys.BuiltInKeys()

	entries := make([]gen.FactKeyEntry, 0, len(builtInKeys))
	for _, key := range builtInKeys {
		builtin := true
		entry := gen.FactKeyEntry{
			Key:     key,
			Builtin: &builtin,
		}
		if desc, ok := builtInDescriptions[key]; ok {
			entry.Description = &desc
		}
		entries = append(entries, entry)
	}

	return gen.GetFactKeys200JSONResponse{
		Keys: entries,
	}, nil
}
