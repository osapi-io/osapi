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

package password

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

// debian sets the password with chpasswd. It asks the user behavior whether
// the user exists.
type debian struct {
	exec  host.Exec
	users user.Provider
}

// newDebian is the provider on the Debian family. It declines a host where
// the user behavior has no provider, and the method reports skipped there.
func newDebian(
	h host.Host,
) (Provider, error) {
	users := host.Require[user.Provider](h, user.Behavior)
	if users == nil {
		return nil, behavior.ErrUnsupported
	}

	return &debian{exec: h.Exec, users: users}, nil
}

// Set sets the password of the user from hash, on the standard input of
// chpasswd, so the hash is in no argument list.
func (d *debian) Set(
	ctx context.Context,
	name string,
	hash string,
) error {
	if hash == "" {
		return fmt.Errorf("no password for user %s: give the password", name)
	}
	if err := checkHash(hash); err != nil {
		return err
	}
	if _, err := behavior.Call(
		ctx, user.Behavior, user.Get, d.users, user.GetInput{Name: name},
	); err != nil {
		return err
	}
	if _, err := d.exec.RunPrivilegedCmdWithStdin(
		ctx, "chpasswd", []string{"-e"}, name+":"+hash+"\n",
	); err != nil {
		return fmt.Errorf("set the password of user %s: %w", name, err)
	}

	return nil
}

// checkHash refuses a hash that is not a crypt hash, or that would break
// the line that chpasswd reads.
func checkHash(
	hash string,
) error {
	if !strings.HasPrefix(hash, "$") || strings.ContainsAny(hash, ":\r\n") {
		return errors.New("the password hash is not a crypt hash")
	}

	return nil
}
