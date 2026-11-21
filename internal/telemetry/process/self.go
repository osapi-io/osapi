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

package process

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/process"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

// provider reads this process with gohai's process collector, which
// reads one PID and not the others.
type provider struct {
	pid     int32
	sources source.Sources
	now     func() time.Time

	mu      sync.Mutex
	last    float64   // cpu_seconds of the last sample
	lastAt  time.Time // when the last sample was read
	sampled bool
}

// New returns a Provider for the current process on the real host.
func New() Provider {
	return &provider{pid: int32(os.Getpid()), now: time.Now}
}

// GetMetrics is this process's CPU share, RSS and goroutine count. The CPU
// share is the change of gohai's cpu_seconds since the last call, over the
// time between them, so a spike shows in the next heartbeat. The first call
// has no sample before it and uses gohai's cpu_percent.
func (p *provider) GetMetrics() (*Metrics, error) {
	proc, err := process.NewWith(p.sources).Read(context.Background(), p.pid)
	if err != nil {
		return nil, fmt.Errorf("get process: %w", err)
	}
	now := p.now()

	p.mu.Lock()
	cpu := proc.CPUPercent
	if elapsed := now.Sub(p.lastAt).Seconds(); p.sampled && elapsed > 0 {
		cpu = (proc.CPUSeconds - p.last) / elapsed * 100
	}
	p.last, p.lastAt, p.sampled = proc.CPUSeconds, now, true
	p.mu.Unlock()

	return &Metrics{
		CPUPercent: cpu,
		RSSBytes:   proc.RSS,
		Goroutines: runtime.NumGoroutine(),
	}, nil
}
