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

// GetInput is get's input, which has none.
type GetInput struct{}

// Entry is the host's load averages.
type Entry struct {
	LoadAverage Averages `json:"load_average" doc:"The load averages of the host"`
}

// Averages are the load averages over 1, 5 and 15 minutes.
type Averages struct {
	One     float64 `json:"1min"  label:"1 min"  doc:"The load average over 1 minute"`
	Five    float64 `json:"5min"  label:"5 min"  doc:"The load average over 5 minutes"`
	Fifteen float64 `json:"15min" label:"15 min" doc:"The load average over 15 minutes"`
}
