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

package os

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// darwin reads the system's name and product version from sysctl.
type darwin struct {
	exec host.Exec
}

// newDarwin is the provider on macOS.
func newDarwin(
	h host.Host,
) (Provider, error) {
	return &darwin{exec: h.Exec}, nil
}

// Release is kern.ostype in lower case, darwin, and kern.osproductversion,
// such as 15.5.
func (d *darwin) Release(
	ctx context.Context,
) (Release, error) {
	var v [2]string
	for i, name := range []string{"kern.ostype", "kern.osproductversion"} {
		out, err := d.exec.RunCmd(ctx, "sysctl", []string{"-n", name})
		if err != nil {
			return Release{}, fmt.Errorf("read %s: %w", name, err)
		}
		v[i] = strings.TrimSpace(out)
	}

	return Release{Distribution: strings.ToLower(v[0]), Version: v[1]}, nil
}
