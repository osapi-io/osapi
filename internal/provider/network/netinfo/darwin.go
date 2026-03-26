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

package netinfo

import (
	"io"
	"net"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
	"github.com/retr0h/osapi/internal/provider"
)

var _ provider.FactsSetter = (*Darwin)(nil)

// Darwin implements the Provider interface for macOS systems.
type Darwin struct {
	Netinfo
	RouteReaderFn func() (io.ReadCloser, error)
}

// NewDarwinProvider factory to create a new Darwin instance.
func NewDarwinProvider(
	em exec.Manager,
) *Darwin {
	return &Darwin{
		Netinfo: Netinfo{
			InterfacesFn: net.Interfaces,
			AddrsFn: func(iface net.Interface) ([]net.Addr, error) {
				return iface.Addrs()
			},
		},
		RouteReaderFn: func() (io.ReadCloser, error) {
			output, err := em.RunCmd("netstat", []string{"-rn"})
			if err != nil {
				return nil, err
			}
			return io.NopCloser(strings.NewReader(output)), nil
		},
	}
}
