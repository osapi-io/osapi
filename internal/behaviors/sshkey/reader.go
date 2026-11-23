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

package sshkey

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/ssh"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

// reader reads the keys through gohai's ssh collector, which reads the
// authorized_keys of each user that /etc/passwd lists, or on macOS the
// directory service. It asks the user behavior whether the user exists. It
// changes nothing.
type reader struct {
	src   source.Sources
	users user.Provider
}

// newReader is the provider on each platform but the Debian family.
func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources(), users: host.Require[user.Provider](h, user.Behavior)}, nil
}

// Keys is the keys of the user, in the order of its file.
func (r *reader) Keys(
	ctx context.Context,
	name string,
) ([]Key, error) {
	if _, err := r.user(ctx, name); err != nil {
		return nil, err
	}

	return keysOf(ctx, r.src, name)
}

// collectSSH runs gohai's ssh collector. A test replaces it to fail: the
// collector of gohai v1.12.1 has no error of its own.
var collectSSH = func(
	ctx context.Context,
	src source.Sources,
) (any, error) {
	return ssh.NewWith(src).Collect(ctx, nil)
}

// keysOf is the keys of name that gohai reads from its
// authorized_keys, in the order of the file.
func keysOf(
	ctx context.Context,
	src source.Sources,
	name string,
) ([]Key, error) {
	out, err := collectSSH(ctx, src)
	if err != nil {
		return nil, err
	}
	keys := []Key{}
	if info, ok := out.(*ssh.Info); ok {
		for _, k := range info.AuthorizedKeys[name] {
			keys = append(keys, Key{
				Type:        k.Type,
				Fingerprint: k.FingerprintSHA256,
				Comment:     k.Comment,
				Options:     k.Options,
			})
		}
	}

	return keys, nil
}

// user reads the user through the user behavior, as its own step.
func (r *reader) user(
	ctx context.Context,
	name string,
) (user.Account, error) {
	return behavior.Call(ctx, user.Behavior, user.Get, r.users, user.GetInput{Name: name})
}

// Add changes nothing off the Debian family.
func (r *reader) Add(
	context.Context,
	string,
	string,
) (string, bool, error) {
	return "", false, behavior.ErrUnsupported
}

// Remove changes nothing off the Debian family.
func (r *reader) Remove(
	context.Context,
	string,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// errNotKey is a line that has no public key.
var errNotKey = errors.New(
	"not an SSH public key: give a line such as ssh-ed25519 AAAA... alice@laptop",
)

// fingerprint is the SHA256 fingerprint of the key of an authorized_keys
// line, as ssh-keygen -l prints it, after any options. A line with no key,
// a comment or an empty line, has none.
func fingerprint(
	line string,
) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", errNotKey
	}
	f := strings.Fields(line)
	for i := 0; i+1 < len(f); i++ {
		if !isKeyType(f[i]) {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(f[i+1])
		if err != nil {
			return "", fmt.Errorf("%w: %w", errNotKey, err)
		}
		sum := sha256.Sum256(raw)

		return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:]), nil
	}

	return "", errNotKey
}

// isKeyType reports whether s names a key type, as gohai's parser does.
func isKeyType(
	s string,
) bool {
	for _, p := range []string{"ssh-", "ecdsa-", "sk-"} {
		if strings.HasPrefix(s, p) {
			return true
		}
	}

	return false
}
