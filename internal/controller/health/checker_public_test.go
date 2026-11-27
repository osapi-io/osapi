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

package health_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/health"
)

type CheckerPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *CheckerPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *CheckerPublicTestSuite) TestCheckHealth() {
	tests := []struct {
		name         string
		checker      *health.NATSChecker
		validateFunc func(error)
	}{
		{
			name: "all checks pass",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return nil },
				KVCheck:   func(context.Context) error { return nil },
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "NATS check fails",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return fmt.Errorf("nats error") },
				KVCheck:   func(context.Context) error { return nil },
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "nats error")
			},
		},
		{
			name: "KV check fails",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return nil },
				KVCheck:   func(context.Context) error { return fmt.Errorf("kv error") },
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "kv error")
			},
		},
		{
			name: "both checks fail",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return fmt.Errorf("nats error") },
				KVCheck:   func(context.Context) error { return fmt.Errorf("kv error") },
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "nats error")
			},
		},
		{
			name:    "nil checks pass",
			checker: &health.NATSChecker{},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tt.checker.CheckHealth(s.ctx))
		})
	}
}

func (s *CheckerPublicTestSuite) TestCheckNATS() {
	tests := []struct {
		name         string
		checker      *health.NATSChecker
		validateFunc func(error)
	}{
		{
			name: "NATS check passes",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return nil },
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "NATS check fails",
			checker: &health.NATSChecker{
				NATSCheck: func(context.Context) error { return fmt.Errorf("nats error") },
			},
			validateFunc: func(err error) {
				s.Error(err)
			},
		},
		{
			name:    "nil NATS check passes",
			checker: &health.NATSChecker{},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tt.checker.CheckNATS(s.ctx))
		})
	}
}

func (s *CheckerPublicTestSuite) TestCheckKV() {
	tests := []struct {
		name         string
		checker      *health.NATSChecker
		validateFunc func(error)
	}{
		{
			name: "KV check passes",
			checker: &health.NATSChecker{
				KVCheck: func(context.Context) error { return nil },
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "KV check fails",
			checker: &health.NATSChecker{
				KVCheck: func(context.Context) error { return fmt.Errorf("kv error") },
			},
			validateFunc: func(err error) {
				s.Error(err)
			},
		},
		{
			name:    "nil KV check passes",
			checker: &health.NATSChecker{},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tt.checker.CheckKV(s.ctx))
		})
	}
}

func (s *CheckerPublicTestSuite) TestChecksGetABoundedContext() {
	type ctxKey struct{}

	tests := []struct {
		name         string
		run          func(c *health.NATSChecker, ctx context.Context) error
		validateFunc func(got context.Context)
	}{
		{
			name: "the NATS check gets the caller's context with a deadline",
			run: func(c *health.NATSChecker, ctx context.Context) error {
				return c.CheckNATS(ctx)
			},
			validateFunc: func(got context.Context) {
				_, ok := got.Deadline()
				s.True(ok, "the check has a deadline")
				s.Equal("caller", got.Value(ctxKey{}))
			},
		},
		{
			name: "the KV check gets the caller's context with a deadline",
			run: func(c *health.NATSChecker, ctx context.Context) error {
				return c.CheckKV(ctx)
			},
			validateFunc: func(got context.Context) {
				_, ok := got.Deadline()
				s.True(ok, "the check has a deadline")
				s.Equal("caller", got.Value(ctxKey{}))
			},
		},
		{
			name: "a cancel of the caller stops the check",
			run: func(c *health.NATSChecker, ctx context.Context) error {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()

				return c.CheckHealth(cancelled)
			},
			validateFunc: func(got context.Context) {
				s.ErrorIs(got.Err(), context.Canceled)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var got context.Context
			record := func(ctx context.Context) error {
				got = ctx

				return ctx.Err()
			}
			c := &health.NATSChecker{NATSCheck: record, KVCheck: record}

			_ = tt.run(c, context.WithValue(s.ctx, ctxKey{}, "caller"))

			tt.validateFunc(got)
		})
	}
}

func TestCheckerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CheckerPublicTestSuite))
}
