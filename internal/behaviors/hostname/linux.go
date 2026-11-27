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

	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// linux reads the kernel's hostname and sets it with hostnamectl.
type linux struct {
	src       source.Sources
	exec      host.Exec
	container bool
}

// newLinux is the provider on Linux.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{src: h.Sources(), exec: h.Exec, container: h.Container}, nil
}

// Hostname is the host's name, read with gohai.
func (l *linux) Hostname(
	ctx context.Context,
) (string, error) {
	return machineName(ctx, l.src)
}

// SetHostname runs hostnamectl set-hostname when gohai reads a different
// name. A container's runtime owns its hostname, so there it is
// unsupported.
func (l *linux) SetHostname(
	ctx context.Context,
	name string,
) (bool, error) {
	if l.container {
		return false, behavior.ErrUnsupported
	}
	current, err := machineName(ctx, l.src)
	if err != nil {
		return false, err
	}
	if current == name {
		return false, nil
	}
	if _, err := l.exec.RunPrivilegedCmd(
		ctx,
		"hostnamectl",
		[]string{"set-hostname", name},
	); err != nil {
		return false, fmt.Errorf("set the hostname: %w", err)
	}

	return true, nil
}
