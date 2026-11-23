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

package managedfile

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// bsdStat are the arguments of the BSD stat of macOS that print
// owner:group. It has no -c.
var bsdStat = []string{"-f", "%Su:%Sg"}

// newDarwin is the provider on macOS: the Linux one, with BSD stat. It
// needs the file state store, and declines a host without one.
func newDarwin(
	h host.Host,
) (Provider, error) {
	if h.FileState == nil {
		return nil, behavior.ErrUnsupported
	}

	return &debian{
		fs:       h.FS,
		state:    h.FileState,
		bucket:   h.FileStateBucket,
		exec:     h.Exec,
		hostname: h.Hostname,
		stat:     bsdStat,
	}, nil
}
