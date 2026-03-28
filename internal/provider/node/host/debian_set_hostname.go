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

package host

import (
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/provider"
	"github.com/retr0h/osapi/pkg/sdk/platform"
)

// SetHostname sets the system hostname using hostnamectl.
// Returns ErrUnsupported in container environments where hostname
// is managed by the container runtime. Checks the current hostname
// first and returns Changed: false if already set to the requested value.
func (u *Debian) SetHostname(
	name string,
) (*SetHostnameResult, error) {
	if platform.IsContainer() {
		return nil, fmt.Errorf("host: %w", provider.ErrUnsupported)
	}

	current, err := u.execManager.RunCmd("hostnamectl", []string{"hostname"})
	if err != nil {
		return nil, fmt.Errorf("failed to get current hostname: %w", err)
	}

	if strings.TrimSpace(current) == name {
		return &SetHostnameResult{Changed: false}, nil
	}

	if _, err := u.execManager.RunCmd("hostnamectl", []string{"set-hostname", name}); err != nil {
		return nil, fmt.Errorf("failed to set hostname: %w", err)
	}

	return &SetHostnameResult{Changed: true}, nil
}
