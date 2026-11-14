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

package behavior

import "strings"

// SpecPath is where OpenAPI's document is written, relative to the
// repository root. While behaviors run beside the hand-written domains it is
// under next/, which redocly join does not read, so the hand-written spec and
// everything generated from it are untouched. The docs site builds its
// API (next) pages from it. internal/behaviorcmd builds the document.
const SpecPath = "internal/controller/api/next/gen/api.yaml"

// TypeName is the prefix of every schema, parameter and operation name the
// spec gives this behavior: sysctl is Sysctl, dns under network is
// NetworkDns.
func (b *Behavior) TypeName() string {
	return title(b.Group) + title(b.Name)
}

// OperationID is the method's operationId, by the rule the hand-written
// specs follow: the HTTP method, the placement, the behavior, and ByKey on a
// keyed read so it does not collide with the list.
func (b *Behavior) OperationID(
	m *Method,
) string {
	id := title(strings.ToLower(m.HTTPMethod())) + title(string(b.Placement)) + b.TypeName()
	if m.Verb == VerbRead && m.KeyField != "" {
		id += "ByKey"
	}

	return id
}

// Tag is the spec tag the docs site groups the behavior's operations under:
// sysctl_operations, or network_dns_operations for a grouped one.
func (b *Behavior) Tag() string {
	if b.Group != "" {
		return b.Group + "_" + b.Name + "_operations"
	}

	return b.Name + "_operations"
}

func title(
	s string,
) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
