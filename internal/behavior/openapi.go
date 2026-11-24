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

// TypeName is the prefix of every schema, parameter and operation name the
// spec gives this behavior: sysctl is Sysctl, and a route under the
// interface is InterfaceRoute.
func (b *Behavior) TypeName() string {
	if b.Parent != nil {
		return b.Parent.TypeName() + title(b.Name)
	}

	return title(b.Name)
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
	if m.Action {
		id += title(m.Name)
	}

	return id
}

// Tag is the spec tag the docs site groups the behavior's operations under,
// such as sysctl_operations.
func (b *Behavior) Tag() string {
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
