// Copyright (c) 2025 John Dewey

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

package job

import (
	"context"
	"time"
)

// waitKey is the context key of a job's wait.
type waitKey struct{}

// WithWait gives the job sent with ctx a wait of d for its response, when d
// is longer than the job client's own timeout: a behavior whose job takes
// longer than that, such as a netplan change, which netplan try confirms
// only after it checks NATS, declares it.
func WithWait(
	ctx context.Context,
	d time.Duration,
) context.Context {
	return context.WithValue(ctx, waitKey{}, d)
}

// Wait is the wait of the job sent with ctx, or timeout when ctx gives a
// shorter one or none.
func Wait(
	ctx context.Context,
	timeout time.Duration,
) time.Duration {
	if d, ok := ctx.Value(waitKey{}).(time.Duration); ok && d > timeout {
		return d
	}

	return timeout
}
