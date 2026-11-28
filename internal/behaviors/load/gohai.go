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

package load

import (
	"context"

	gohai "github.com/osapi-io/gohai/pkg/gohai/collectors/load"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the load averages with gohai's load collector, on Linux and
// on macOS, through the host's sources.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Load is the averages of the run queue over 1, 5 and 15 minutes.
func (r *reader) Load(
	ctx context.Context,
) (Averages, error) {
	got, err := gohai.NewWith(r.src).Collect(ctx, nil)
	l, err := host.Collected[gohai.Info](got, err, "read load")
	if err != nil {
		return Averages{}, err
	}

	return Averages{One: l.One, Five: l.Five, Fifteen: l.Fifteen}, nil
}
