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

package types

import "slices"

// Interval is how often a scheduled job runs, in place of a cron schedule.
type Interval string

// The values of Interval: the intervals of a scheduled job.
const (
	// IntervalHourly runs the job each hour.
	IntervalHourly Interval = "hourly"
	// IntervalDaily runs the job each day.
	IntervalDaily Interval = "daily"
	// IntervalWeekly runs the job each week.
	IntervalWeekly Interval = "weekly"
	// IntervalMonthly runs the job each month.
	IntervalMonthly Interval = "monthly"
)

// Values are the intervals of a scheduled job.
func (Interval) Values() []Interval {
	return []Interval{IntervalHourly, IntervalDaily, IntervalWeekly, IntervalMonthly}
}

// Valid reports whether i is one of the values.
func (i Interval) Valid() bool {
	return slices.Contains(i.Values(), i)
}
