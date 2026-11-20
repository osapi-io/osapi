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
	"fmt"
	"strconv"
	"strings"
)

// averages is the first three numbers of text, such as /proc/loadavg's
// "0.32 0.28 0.25 1/412 12345" or sysctl's "{ 0.32 0.28 0.25 }".
func averages(
	text string,
) (Averages, error) {
	fields := strings.Fields(strings.Trim(strings.TrimSpace(text), "{}"))
	if len(fields) < 3 {
		return Averages{}, fmt.Errorf("no load averages in %q", text)
	}
	var v [3]float64
	for i := range v {
		n, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return Averages{}, fmt.Errorf("load average %q: %w", fields[i], err)
		}
		v[i] = n
	}

	return Averages{One: v[0], Five: v[1], Fifteen: v[2]}, nil
}
