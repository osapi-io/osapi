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

package vault

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// ErrNotFound is the error of a key that a vault does not have.
var ErrNotFound = errors.New("secret not found")

// keyPattern is a key that each type can store: letters, digits, - and _.
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// CheckKey is an error when key is not a name that a vault takes: 1 to 128
// letters, digits, - or _, such as registry-token.
func CheckKey(
	key string,
) error {
	if !keyPattern.MatchString(key) {
		return fmt.Errorf(
			"key %q: a key is 1 to 128 letters, digits, - or _, such as registry-token",
			key,
		)
	}

	return nil
}

// Set is the vaults of the controller, by name, each with its spec.
type Set struct {
	vaults map[string]Vault
	specs  map[string]Spec
	names  []string
}

// Open makes a vault of each spec, from its registered type, in the order
// of specs. A spec with a type that is not registered, a config block that
// does not decode, a check that fails or a vault that does not open is an
// error that names the vault.
func Open(
	ctx context.Context,
	env Env,
	specs []Spec,
) (*Set, error) {
	s := &Set{vaults: map[string]Vault{}, specs: map[string]Spec{}}
	for _, spec := range specs {
		if _, ok := s.vaults[spec.Name]; ok {
			return nil, fmt.Errorf("vault %q: two vaults have the name", spec.Name)
		}
		t, err := TypeOf(spec)
		if err != nil {
			return nil, err
		}
		cfg, err := t.Decode(spec)
		if err != nil {
			return nil, err
		}
		if err := t.Check(spec); err != nil {
			return nil, err
		}
		v, err := t.build(ctx, env, spec, reflect.ValueOf(cfg).Elem().Interface())
		if err != nil {
			return nil, fmt.Errorf("vault %q: %w", spec.Name, err)
		}
		s.vaults[spec.Name] = v
		s.specs[spec.Name] = spec
		s.names = append(s.names, spec.Name)
	}

	return s, nil
}

// Get is the vault name.
func (s *Set) Get(
	name string,
) (Vault, bool) {
	v, ok := s.vaults[name]

	return v, ok
}

// Names is the name of each vault, in the order of the config.
func (s *Set) Names() []string {
	return s.names
}

// Spec is the spec of the vault name, with its type and its targets.
func (s *Set) Spec(
	name string,
) (Spec, bool) {
	spec, ok := s.specs[name]

	return spec, ok
}
