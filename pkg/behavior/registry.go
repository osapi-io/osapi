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
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]*Behavior{}
)

// Register adds b to the registry and returns it, so a behavior's package
// can declare and register in one statement:
//
//	var Behavior = behavior.Register(behavior.Define("sysctl", ...))
//
// A behavior is registered when its package is imported, and
// pkg/behavior/all imports every one. Registering two behaviors with the same
// name keeps the last; behavior validate reports the clash.
func Register(
	b *Behavior,
) *Behavior {
	mu.Lock()
	defer mu.Unlock()

	if prev, ok := registry[b.Name]; ok && prev != b {
		duplicates = append(duplicates, b.Name)
	}
	registry[b.Name] = b

	return b
}

// duplicates records names registered more than once, for Validate.
var duplicates []string

// All returns every registered behavior, sorted by group and name.
func All() []*Behavior {
	mu.RLock()
	defer mu.RUnlock()

	out := make([]*Behavior, 0, len(registry))
	for _, b := range registry {
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}

		return out[i].Name < out[j].Name
	})

	return out
}

// Lookup returns the behavior with the given name, or nil.
func Lookup(
	name string,
) *Behavior {
	mu.RLock()
	defer mu.RUnlock()

	return registry[name]
}

// Reset empties the registry. Tests use it to start from nothing.
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	registry = map[string]*Behavior{}
	duplicates = nil
}
