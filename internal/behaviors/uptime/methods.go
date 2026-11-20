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

package uptime

import (
	"context"
	"fmt"
	"time"
)

// Get is how long the host has run.
func Get(
	ctx context.Context,
	p Provider,
	_ GetInput,
) (Entry, error) {
	d, err := p.Uptime(ctx)
	if err != nil {
		return Entry{}, err
	}

	return Entry{Uptime: words(d), Seconds: int(d.Seconds())}, nil
}

// words is d in days, hours and minutes: 1 day, 2 hours, 5 minutes.
func words(
	d time.Duration,
) string {
	minutes := int(d.Minutes())

	return fmt.Sprintf("%s, %s, %s",
		count(minutes/(24*60), "day"),
		count(minutes%(24*60)/60, "hour"),
		count(minutes%60, "minute"))
}

// count is n and the unit, plural unless n is 1.
func count(
	n int,
	unit string,
) string {
	if n == 1 {
		return "1 " + unit
	}

	return fmt.Sprintf("%d %ss", n, unit)
}
