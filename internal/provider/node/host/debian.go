// Copyright (c) 2024 John Dewey

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

package host

import (
	iexec "github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/provider"
)

var _ provider.FactsSetter = (*Debian)(nil)

// Debian implements the Host Provider interface for Debian and Ubuntu. It
// embeds Linux, whose reads are the same on each distribution, and adds what
// is specific to Debian: UpdateHostname, with hostnamectl.
type Debian struct {
	Linux

	execManager iexec.Manager
}

// NewDebianProvider factory to create a new Debian instance.
func NewDebianProvider(
	execManager iexec.Manager,
) *Debian {
	return &Debian{
		Linux:       *NewLinuxProvider(),
		execManager: execManager,
	}
}

// ExecNotFoundError wraps exec.ErrNotFound for testability.
type ExecNotFoundError struct {
	Name string
}

// Error implements the error interface.
func (e *ExecNotFoundError) Error() string {
	return "executable file not found: " + e.Name
}
