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

// Package settings is the registry of who owns each section of osapi's
// config and its defaults. A DSL definition declares the defaults of the
// values it owns (command.Config for a process, behavior.Config for a
// behavior), and package config reads them from here to make each key. It
// imports nothing of osapi, so a DSL and config both import it.
package settings

import (
	"fmt"
	"sort"
	"sync"
)

// Owner is the definition that owns a section of the config.
type Owner struct {
	// Prefix is the section, such as controller.
	Prefix string
	// By names the definition, such as controller start.
	By string
	// Defaults is a value of the section's type whose fields are the
	// defaults of its keys.
	Defaults any
}

var (
	mu     sync.Mutex
	owners = map[string]Owner{}
)

// Own records that by owns the section prefix, with defaults. Two owners of
// one section is a mistake of the code, so it panics, as a group registered
// twice does.
func Own(
	prefix string,
	by string,
	defaults any,
) {
	mu.Lock()
	defer mu.Unlock()
	if o, ok := owners[prefix]; ok {
		panic(fmt.Sprintf("settings: %s and %s both own %q", o.By, by, prefix))
	}
	owners[prefix] = Owner{Prefix: prefix, By: by, Defaults: defaults}
}

// Of is the owner of the section prefix.
func Of(
	prefix string,
) (Owner, bool) {
	mu.Lock()
	defer mu.Unlock()
	o, ok := owners[prefix]

	return o, ok
}

// All is each owner, by its section.
func All() []Owner {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Owner, 0, len(owners))
	for _, o := range owners {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Prefix < out[j].Prefix })

	return out
}
