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

// Package vault is the store of secrets that the controller resolves a
// secret://<vault>/<key> reference from. A vault has a name and a type. A
// type registers with its name, a constructor and the Go struct of its
// config block, so a different store, such as HashiCorp Vault, is a type
// that plugs in. osapi ships one type, nats, in package natsvault.
package vault

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
)

var (
	mu    sync.RWMutex
	types = map[string]Type{}
)

// Define is a type of vault: name, what it is in one line, and build, which
// makes a vault from its spec and its config block as a C. check, when it
// is not nil, checks the spec of a vault of the type, such as that it names
// its key. The config check runs it, and Open runs it before build.
func Define[C any](
	name string,
	doc string,
	check func(spec Spec) error,
	build func(ctx context.Context, env Env, spec Spec, cfg C) (Vault, error),
) Type {
	return Type{
		Name:   name,
		Doc:    doc,
		Config: reflect.TypeFor[C](),
		check:  check,
		build: func(ctx context.Context, env Env, spec Spec, cfg any) (Vault, error) {
			return build(ctx, env, spec, cfg.(C))
		},
	}
}

// Register adds t to the types that the controller and the config check
// know, and returns it. Two types with one name are a mistake in the code,
// so it panics.
func Register(
	t Type,
) Type {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := types[t.Name]; ok {
		panic(fmt.Sprintf("vault: type %q registered twice", t.Name))
	}
	types[t.Name] = t

	return t
}

// Unregister removes the type name. A test that registers a type removes
// it with this.
func Unregister(
	name string,
) {
	mu.Lock()
	defer mu.Unlock()
	delete(types, name)
}

// Lookup is the registered type name.
func Lookup(
	name string,
) (Type, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := types[name]

	return t, ok
}

// Types is each registered type, sorted by name.
func Types() []Type {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Type, 0, len(types))
	for _, t := range types {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

// TypeOf is the registered type of spec, or an error that names the types
// that there are.
func TypeOf(
	spec Spec,
) (Type, error) {
	t, ok := Lookup(spec.Type)
	if ok {
		return t, nil
	}
	names := make([]string, 0, len(types))
	for _, t := range Types() {
		names = append(names, t.Name)
	}

	return Type{}, fmt.Errorf(
		"vault %q: type %q is not a type of vault: the types are %s",
		spec.Name, spec.Type, strings.Join(names, ", "),
	)
}

// Decode is the config block of spec as the struct of t: a pointer to a
// new value of t.Config. A key that the struct does not have is an error,
// so a typo does not pass.
func (t Type) Decode(
	spec Spec,
) (any, error) {
	out := reflect.New(t.Config)
	var md mapstructure.Metadata
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         &md,
		WeaklyTypedInput: true,
		Result:           out.Interface(),
	})
	if err := dec.Decode(spec.Config); err != nil {
		return nil, fmt.Errorf("vault %q: config: %w", spec.Name, err)
	}
	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)

		return nil, fmt.Errorf(
			"vault %q: config: the type %s has no setting %s",
			spec.Name, t.Name, strings.Join(md.Unused, ", "),
		)
	}

	return out.Interface(), nil
}

// Check is the check of the type on spec, or nil for a type with none.
func (t Type) Check(
	spec Spec,
) error {
	if t.check == nil {
		return nil
	}
	if err := t.check(spec); err != nil {
		return fmt.Errorf("vault %q: %w", spec.Name, err)
	}

	return nil
}
