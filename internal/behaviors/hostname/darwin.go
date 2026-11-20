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

package hostname

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// darwin reads kern.hostname. It does not set the name: update runs only on
// debian.
type darwin struct {
	exec host.Exec
}

// newDarwin is the provider on macOS.
func newDarwin(
	h host.Host,
) (Provider, error) {
	return &darwin{exec: h.Exec}, nil
}

// Hostname is kern.hostname.
func (d *darwin) Hostname(
	ctx context.Context,
) (string, error) {
	out, err := d.exec.RunCmd(ctx, "sysctl", []string{"-n", "kern.hostname"})
	if err != nil {
		return "", fmt.Errorf("read kern.hostname: %w", err)
	}

	return strings.TrimSpace(out), nil
}

// SetHostname is unsupported on macOS. The agent does not call it, because
// update runs only on debian.
func (d *darwin) SetHostname(
	context.Context,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}
