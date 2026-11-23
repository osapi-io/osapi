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
	"context"
	"fmt"
	"time"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/hostname"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/platform"
	"github.com/osapi-io/gohai/pkg/gohai/collectors/uptime"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

var _ Provider = (*Gohai)(nil)

// Gohai reads the host's name, OS and uptime with gohai's collectors, on
// every Linux family and on macOS.
type Gohai struct {
	// Sources are what the collectors read. A zero field is the real
	// host's.
	Sources source.Sources
}

// New returns a provider that reads the real host.
func New() *Gohai {
	return &Gohai{}
}

// GetHostname is the host's name, as `hostname` prints it.
func (g *Gohai) GetHostname() (string, error) {
	got, err := hostname.NewWith(g.Sources).Collect(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("get hostname: %w", err)
	}

	return got.(*hostname.Info).MachineName, nil
}

// GetOSInfo is the distribution and its version, as gohai's platform
// collector names them: ubuntu 24.04, darwin 15.4.
func (g *Gohai) GetOSInfo() (*Result, error) {
	// The platform collector reads what it can, and does not fail.
	got, _ := platform.NewWith(g.Sources).Collect(context.Background(), nil)
	info, _ := got.(*platform.Info)

	return &Result{Distribution: info.Name, Version: info.Version}, nil
}

// GetUptime is the time since the host booted.
func (g *Gohai) GetUptime() (time.Duration, error) {
	got, err := uptime.NewWith(g.Sources).Collect(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("get uptime: %w", err)
	}

	return time.Duration(got.(*uptime.Info).Seconds) * time.Second, nil
}
