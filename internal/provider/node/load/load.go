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

// Package load reads the load averages of the host.
package load

import (
	"context"
	"fmt"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/load"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

var _ Provider = (*Gohai)(nil)

// Gohai reads the host's load averages with gohai's load collector, on
// every Linux family and on macOS.
type Gohai struct {
	// Sources are what the collector reads. A zero field is the real
	// host's.
	Sources source.Sources
}

// New returns a provider that reads the real host.
func New() *Gohai {
	return &Gohai{}
}

// GetAverageStats is the 1, 5 and 15 minute load averages.
func (g *Gohai) GetAverageStats() (*Result, error) {
	got, err := load.NewWith(g.Sources).Collect(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("get load: %w", err)
	}
	l := got.(*load.Info)

	return &Result{Load1: float32(l.One), Load5: float32(l.Five), Load15: float32(l.Fifteen)}, nil
}
