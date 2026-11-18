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
)

// NATSChecker checks NATS and KV bucket connectivity.
type NATSChecker struct {
	// NATSCheck verifies NATS connectivity.
	NATSCheck func() error
	// KVCheck verifies KV bucket accessibility.
	KVCheck func() error
}

// CheckHealth runs all dependency checks and returns the first error.
func (c *NATSChecker) CheckHealth(
	_ context.Context,
) error {
	var errs []error

	if c.NATSCheck != nil {
		if err := c.NATSCheck(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.KVCheck != nil {
		if err := c.KVCheck(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// CheckNATS runs only the NATS connectivity check.
func (c *NATSChecker) CheckNATS() error {
	if c.NATSCheck != nil {
		return c.NATSCheck()
	}

	return nil
}

// CheckKV runs only the KV bucket check.
func (c *NATSChecker) CheckKV() error {
	if c.KVCheck != nil {
		return c.KVCheck()
	}

	return nil
}
