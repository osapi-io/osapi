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

import (
	"sync"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	groupsMu sync.RWMutex
	groups   = map[string]*endpoint.Group{}
)

// RegisterGroup records g as the endpoint group of the behavior with the
// given name. The generated gen/group.go of each behavior with an API calls
// it, and the endpoint layer makes the behavior's spec from the group.
func RegisterGroup(
	name string,
	g *endpoint.Group,
) {
	groupsMu.Lock()
	defer groupsMu.Unlock()

	groups[name] = g
}

// GroupOf is b's endpoint group, or nil when its gen package was not
// imported or b has no API.
func GroupOf(
	b *Behavior,
) *endpoint.Group {
	groupsMu.RLock()
	defer groupsMu.RUnlock()

	return groups[b.Name]
}
