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
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
)

// Spec is the config file as a spec: a section for each top word of a key,
// such as controller, and a term for each key, from Keys. osapi config spec
// get prints it, as each DSL's spec get prints the parts of the DSL.
func Spec() guide.Spec {
	return SpecOf(Keys())
}

// SpecOf is the spec of keys.
func SpecOf(
	keys []Key,
) guide.Spec {
	var s guide.Spec
	seen := map[string]bool{}
	for _, k := range keys {
		kind, _, _ := strings.Cut(k.Name, ".")
		if !seen[kind] {
			seen[kind] = true
			s.Kinds = append(s.Kinds, guide.Kind{Kind: kind, Title: kind})
		}
		does := k.Doc
		if does == "" {
			does = "No description."
		}
		does += " Type: " + k.Type + "."
		if k.Default != nil && fmt.Sprint(k.Default) != "" {
			does += fmt.Sprintf(" Default: %v.", k.Default)
		}
		when := "In osapi.yaml, or with " + k.Env + ", which wins over the file."
		if k.Validate != "" {
			when += " Rules: " + k.Validate + "."
		}
		if k.Secret {
			when += " A secret: osapi prints it as ********."
		}
		s.Terms = append(s.Terms, guide.Term{Kind: kind, Name: k.Name, Does: does, When: when})
	}

	return s
}
