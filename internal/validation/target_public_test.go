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

package validation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/validation"
)

type TargetPublicTestSuite struct {
	suite.Suite
}

type targetInput struct {
	Target string `validate:"required,valid_target"`
}

func (s *TargetPublicTestSuite) TestValidTarget() {
	tests := []struct {
		name        string
		setupLister func()
		input       targetInput
		wantOK      bool
		contains    []string
	}{
		{
			name: "when target is _any",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "_any"},
			wantOK: true,
		},
		{
			name: "when target is _all",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "_all"},
			wantOK: true,
		},
		{
			name: "when target is a label with exact match",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "group:web"},
			wantOK: true,
		},
		{
			name: "when target is a label with hierarchical prefix match",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{
								Hostname: "server1",
								Labels:   map[string]string{"group": "web.dev.us-east"},
							},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "group:web.dev"},
			wantOK: true,
		},
		{
			name: "when target label does not match any worker",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "env:prod"},
			wantOK: false,
		},
		{
			name: "when label has empty key",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: ":value"},
			wantOK: false,
		},
		{
			name: "when label has empty value",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "key:"},
			wantOK: false,
		},
		{
			name: "when label is malformed colons",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: ":::"},
			wantOK: false,
		},
		{
			name: "when label key has invalid characters",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "gr@up:web"},
			wantOK: false,
		},
		{
			name: "when label value segment has invalid characters",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "group:web/dev"},
			wantOK: false,
		},
		{
			name: "when target is a known worker hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:  targetInput{Target: "server1"},
			wantOK: true,
		},
		{
			name: "when target is an unknown hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return []validation.WorkerTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:    targetInput{Target: "nonexistent"},
			wantOK:   false,
			contains: []string{"valid_target", "target worker", "nonexistent", "not found"},
		},
		{
			name: "when lister returns error for hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return nil, fmt.Errorf("nats connection failed")
					},
				)
			},
			input:  targetInput{Target: "server1"},
			wantOK: false,
		},
		{
			name: "when lister returns error for label",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.WorkerTarget, error) {
						return nil, fmt.Errorf("nats connection failed")
					},
				)
			},
			input:  targetInput{Target: "group:web"},
			wantOK: false,
		},
		{
			name: "when lister is nil",
			setupLister: func() {
				validation.RegisterTargetValidator(nil)
			},
			input:  targetInput{Target: "server1"},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupLister()

			errMsg, ok := validation.Struct(tt.input)
			s.Equal(tt.wantOK, ok)

			if !ok {
				for _, c := range tt.contains {
					s.Contains(errMsg, c)
				}
			}
		})
	}
}

func TestTargetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TargetPublicTestSuite))
}
