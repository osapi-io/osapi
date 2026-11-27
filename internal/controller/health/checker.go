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

package health

import (
	"context"
	"errors"
	"time"
)

// checkTimeout bounds one check, so a slow NATS can't hold a probe of
// /health/ready, which anyone can call, for longer than this.
const checkTimeout = 5 * time.Second

// NATSChecker checks NATS and KV bucket connectivity.
type NATSChecker struct {
	// NATSCheck verifies NATS connectivity.
	NATSCheck func(ctx context.Context) error
	// KVCheck verifies KV bucket accessibility.
	KVCheck func(ctx context.Context) error
	// StoreCheck says when the jobs bucket is near its max_bytes. It is
	// not part of CheckHealth: a controller with a nearly full store still
	// takes traffic, and frees room by removing finished jobs.
	StoreCheck func(ctx context.Context) error
}

// CheckHealth runs all dependency checks and returns the first error.
func (c *NATSChecker) CheckHealth(
	ctx context.Context,
) error {
	return errors.Join(c.CheckNATS(ctx), c.CheckKV(ctx))
}

// CheckNATS runs only the NATS connectivity check.
func (c *NATSChecker) CheckNATS(
	ctx context.Context,
) error {
	return run(ctx, c.NATSCheck)
}

// CheckKV runs only the KV bucket check.
func (c *NATSChecker) CheckKV(
	ctx context.Context,
) error {
	return run(ctx, c.KVCheck)
}

// CheckStore runs only the job store check.
func (c *NATSChecker) CheckStore(
	ctx context.Context,
) error {
	return run(ctx, c.StoreCheck)
}

// run calls check, when it is set, with ctx bounded by checkTimeout.
func run(
	ctx context.Context,
	check func(ctx context.Context) error,
) error {
	if check == nil {
		return nil
	}

	checkCtx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	return check(checkCtx)
}
