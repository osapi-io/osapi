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
	"regexp"
	"strings"

	"github.com/osapi-io/osapi/internal/secret"
)

// namePattern is the name of a vault, as the rule of Spec.Name has it.
var namePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// Reference names a key in a vault: secret://<vault>/<key>.
type Reference struct {
	Vault string
	Key   string
}

// String is the reference as a config value or an input gives it.
func (r Reference) String() string {
	return secret.Scheme + r.Vault + "/" + r.Key
}

// IsReference reports whether s starts with secret.Scheme. A value that does
// is a reference, and never a literal value.
func IsReference(
	s string,
) bool {
	return strings.HasPrefix(s, secret.Scheme)
}

// ParseReference is the reference s: the name of a vault, then its key.
func ParseReference(
	s string,
) (Reference, error) {
	if !IsReference(s) {
		return Reference{}, fmt.Errorf(
			"%q is not a reference: a reference is %s<vault>/<key>",
			s,
			secret.Scheme,
		)
	}
	name, key, ok := strings.Cut(strings.TrimPrefix(s, secret.Scheme), "/")
	if !ok || !namePattern.MatchString(name) {
		return Reference{}, fmt.Errorf(
			"%s: a reference is %s<vault>/<key>, with a vault name of 1 to 64 letters, digits, - or _",
			s,
			secret.Scheme,
		)
	}
	if err := CheckKey(key); err != nil {
		return Reference{}, fmt.Errorf("%s: %w", s, err)
	}

	return Reference{Vault: name, Key: key}, nil
}

// Resolve is the value of the key that ref names, from the vault of s that
// ref names. An error names the reference, and never a value.
func (s *Set) Resolve(
	ctx context.Context,
	ref Reference,
) ([]byte, error) {
	v, ok := s.Get(ref.Vault)
	if !ok {
		return nil, fmt.Errorf("%s: the vault %q does not exist", ref, ref.Vault)
	}
	value, _, err := v.Get(ctx, ref.Key)
	if errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("%s: the vault %q has no key %q", ref, ref.Vault, ref.Key)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ref, err)
	}

	return value, nil
}
