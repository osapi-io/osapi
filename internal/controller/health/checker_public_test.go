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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/health"
)

// callerKey marks the context of the caller, so a check can tell that it
// got that context.
type callerKey struct{}

type CheckerPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *CheckerPublicTestSuite) SetupTest() {
	s.ctx = context.WithValue(context.Background(), callerKey{}, "caller")
}

// fromCaller is a check that fails unless it got the caller's context
// with a deadline.
func fromCaller(
	ctx context.Context,
) error {
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("the check has no deadline")
	}
	if ctx.Value(callerKey{}) != "caller" {
		return errors.New("the check did not get the caller's context")
	}

	return ctx.Err()
}

func (s *CheckerPublicTestSuite) TestCheckHealth() {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name    string
		checker *health.NATSChecker
		// ctx is the caller's context; nil is the suite's.
		ctx          context.Context
		validateFunc func(error)
	}{
		{
			name:    "each check gets the caller's context, with a deadline",
			checker: &health.NATSChecker{NATSCheck: fromCaller, KVCheck: fromCaller},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:    "a cancel of the caller stops each check",
			checker: &health.NATSChecker{NATSCheck: fromCaller, KVCheck: fromCaller},
			ctx:     context.WithValue(cancelled, callerKey{}, "caller"),
			validateFunc: func(err error) {
				s.ErrorIs(err, context.Canceled)
			},
		},
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
			ctx := tt.ctx
			if ctx == nil {
				ctx = s.ctx
			}
			tt.validateFunc(tt.checker.CheckHealth(ctx))
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
			name:    "the check gets the caller's context, with a deadline",
			checker: &health.NATSChecker{NATSCheck: fromCaller},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
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
			name:    "the check gets the caller's context, with a deadline",
			checker: &health.NATSChecker{KVCheck: fromCaller},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
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

func TestCheckerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CheckerPublicTestSuite))
}
