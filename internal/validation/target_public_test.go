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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/osapi-io/osapi/internal/validation"
)

type TargetPublicTestSuite struct {
	suite.Suite
}

type targetInput struct {
	Target string `validate:"required,valid_target"`
}

func (s *TargetPublicTestSuite) TestValidTarget() {
	tests := []struct {
		name         string
		setupLister  func()
		input        targetInput
		wantOK       bool
		contains     []string
		validateFunc func()
	}{
		{
			name: "when target is _any",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
			name: "when target is _all and all agents pending",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{Hostname: "server1", State: "Pending"},
							{Hostname: "server2", State: "Pending"},
						}, nil
					},
				)
			},
			input:    targetInput{Target: "_all"},
			wantOK:   false,
			contains: []string{"all agents are pending PKI enrollment"},
		},
		{
			name: "when target is _any and all agents pending",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{Hostname: "server1", State: "Pending"},
						}, nil
					},
				)
			},
			input:    targetInput{Target: "_any"},
			wantOK:   false,
			contains: []string{"all agents are pending PKI enrollment"},
		},
		{
			name: "when target is _all and lister returns error",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return nil, fmt.Errorf("nats down")
					},
				)
			},
			input:  targetInput{Target: "_all"},
			wantOK: true, // error → can't determine pending, allow through
		},
		{
			name: "when target is _all and no agents registered",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{}, nil
					},
				)
			},
			input:  targetInput{Target: "_all"},
			wantOK: true, // no agents → can't determine pending, allow through
		},
		{
			name: "when target is _all and lister is nil",
			setupLister: func() {
				validation.RegisterTargetValidator(nil)
			},
			input:  targetInput{Target: "_all"},
			wantOK: true, // nil lister → can't determine pending, allow through
		},
		{
			name: "when target is _all with mix of pending and ready agents",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{Hostname: "server1", State: "Pending"},
							{Hostname: "server2", State: "Ready"},
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
			name: "when target label does not match any agent",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
			name: "when target is a known agent hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
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
			name: "when target matches a pending agent returns false with pending message",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{Hostname: "pending-host", State: "Pending"},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:    targetInput{Target: "pending-host"},
			wantOK:   false,
			contains: []string{"valid_target", "pending PKI enrollment"},
		},
		{
			name: "when target is an unknown hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{Hostname: "server1", Labels: map[string]string{"group": "web"}},
							{Hostname: "server2"},
						}, nil
					},
				)
			},
			input:    targetInput{Target: "nonexistent"},
			wantOK:   false,
			contains: []string{"valid_target", "target agent", "nonexistent", "not found"},
		},
		{
			name: "when lister returns error for hostname",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
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
					func(_ context.Context) ([]validation.AgentTarget, error) {
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
		{
			name: "when same target validated twice uses cache",
			validateFunc: func() {
				callCount := 0
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						callCount++

						return []validation.AgentTarget{
							{Hostname: "server1"},
						}, nil
					},
				)

				// First call populates cache.
				_, ok := validation.Struct(context.Background(), targetInput{Target: "server1"})
				s.True(ok)
				s.Equal(1, callCount)

				// Second call should use cache, not call lister again.
				_, ok = validation.Struct(context.Background(), targetInput{Target: "server1"})
				s.True(ok)
				s.Equal(1, callCount)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.validateFunc != nil {
				tt.validateFunc()

				return
			}

			tt.setupLister()

			errMsg, ok := validation.Struct(context.Background(), tt.input)
			s.Equal(tt.wantOK, ok)

			if !ok {
				for _, c := range tt.contains {
					s.Contains(errMsg, c)
				}
			}
		})
	}
}

// TestValidTargetUsesRequestContext checks that valid_target lists the
// agents with the request's context, so the lookup is in the request's
// trace and a cancelled request stops it. Before #868 it used
// context.Background(), and the lister saw none of the request's values.
func (s *TargetPublicTestSuite) TestValidTargetUsesRequestContext() {
	type key struct{}
	type in struct {
		Target string `validate:"valid_target"`
	}

	tests := []struct {
		name string
		call func(ctx context.Context) bool
	}{
		{
			name: "when a struct is validated the lister gets the request's context",
			call: func(ctx context.Context) bool {
				_, ok := validation.Struct(ctx, in{Target: "web-01"})
				return ok
			},
		},
		{
			name: "when a target is resolved the lister gets the caller's context",
			call: func(ctx context.Context) bool {
				return validation.ResolveTarget(ctx, "web-01") != ""
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got any
			validation.RegisterTargetValidator(
				func(ctx context.Context) ([]validation.AgentTarget, error) {
					got = ctx.Value(key{})
					return []validation.AgentTarget{{MachineID: "m1", Hostname: "web-01"}}, nil
				},
			)

			s.True(tc.call(context.WithValue(context.Background(), key{}, "request")))
			s.Equal("request", got)
		})
	}
}

// TestLookupSpan checks that a registry read by a target check has a span
// that says what it is for, with the target, so a trace tells it from the
// agent.list of a job: validate.target for a rule, target.resolve for a
// hostname. A read from the cache calls nothing, so it has no span.
func (s *TargetPublicTestSuite) TestLookupSpan() {
	type in struct {
		Target string `validate:"valid_target"`
	}

	tests := []struct {
		name   string
		cached bool
		call   func(ctx context.Context)
		want   string
		target string
	}{
		{
			name: "when a rule validates a hostname opens validate.target",
			call: func(ctx context.Context) {
				_, _ = validation.Struct(ctx, in{Target: "web-01"})
			},
			want:   "validate.target",
			target: "web-01",
		},
		{
			name: "when a rule validates _all opens validate.target",
			call: func(ctx context.Context) {
				_, _ = validation.Struct(ctx, in{Target: "_all"})
			},
			want:   "validate.target",
			target: "_all",
		},
		{
			name: "when a hostname is resolved opens target.resolve",
			call: func(ctx context.Context) {
				_ = validation.ResolveTarget(ctx, "web-01")
			},
			want:   "target.resolve",
			target: "web-01",
		},
		{
			name:   "when the agents are cached opens no span",
			cached: true,
			call: func(ctx context.Context) {
				_, _ = validation.Struct(ctx, in{Target: "web-01"})
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			validation.RegisterTargetValidator(
				func(context.Context) ([]validation.AgentTarget, error) {
					return []validation.AgentTarget{{MachineID: "m1", Hostname: "web-01"}}, nil
				},
			)
			if tc.cached {
				_, _ = validation.Struct(context.Background(), in{Target: "web-01"})
			}
			recorder := tracetest.NewSpanRecorder()
			previous := otel.GetTracerProvider()
			otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
			defer otel.SetTracerProvider(previous)

			ctx, request := otel.Tracer("test").
				Start(context.Background(), "GET /api/node/:hostname")
			tc.call(ctx)
			request.End()

			var lookups []sdktrace.ReadOnlySpan
			for _, sp := range recorder.Ended() {
				if sp.Name() != "GET /api/node/:hostname" {
					lookups = append(lookups, sp)
				}
			}
			if tc.want == "" {
				s.Empty(lookups)
				return
			}
			s.Require().Len(lookups, 1)
			s.Equal(tc.want, lookups[0].Name())
			s.Equal(request.SpanContext().SpanID(), lookups[0].Parent().SpanID())
			s.Contains(lookups[0].Attributes(), attribute.String("osapi.target", tc.target))
		})
	}
}

func (s *TargetPublicTestSuite) TestResolveTarget() {
	tests := []struct {
		name         string
		setupLister  func()
		target       string
		validateFunc func(string)
	}{
		{
			name:   "when target is _any returns unchanged",
			target: "_any",
			validateFunc: func(got string) {
				s.Equal("_any", got)
			},
		},
		{
			name:   "when target is _all returns unchanged",
			target: "_all",
			validateFunc: func(got string) {
				s.Equal("_all", got)
			},
		},
		{
			name:   "when target is label returns unchanged",
			target: "group:web",
			validateFunc: func(got string) {
				s.Equal("group:web", got)
			},
		},
		{
			name: "when target is hostname resolves to machine ID",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{MachineID: "abc123", Hostname: "web-01"},
						}, nil
					},
				)
			},
			target: "web-01",
			validateFunc: func(got string) {
				s.Equal("abc123", got)
			},
		},
		{
			name: "when target is machine ID returns unchanged",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{
							{MachineID: "abc123", Hostname: "web-01"},
						}, nil
					},
				)
			},
			target: "abc123",
			validateFunc: func(got string) {
				s.Equal("abc123", got)
			},
		},
		{
			name: "when target not found returns unchanged",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return []validation.AgentTarget{}, nil
					},
				)
			},
			target: "unknown",
			validateFunc: func(got string) {
				s.Equal("unknown", got)
			},
		},
		{
			name: "when lister errors returns unchanged",
			setupLister: func() {
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return nil, fmt.Errorf("nats down")
					},
				)
			},
			target: "web-01",
			validateFunc: func(got string) {
				s.Equal("web-01", got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupLister != nil {
				tt.setupLister()
			}
			got := validation.ResolveTarget(context.Background(), tt.target)
			tt.validateFunc(got)
		})
	}
}

func (s *TargetPublicTestSuite) TestValidTargetMatchesMachineID() {
	validation.RegisterTargetValidator(
		func(_ context.Context) ([]validation.AgentTarget, error) {
			return []validation.AgentTarget{
				{MachineID: "abc123-def456", Hostname: "web-01"},
			}, nil
		},
	)

	// Machine ID should pass validation
	_, ok := validation.Struct(context.Background(), targetInput{Target: "abc123-def456"})
	s.True(ok)
}

// TestValidTargetReadsAgainOnAMiss: a host that registered, or took a new
// name, after the cache was filled is found, because a miss in the cache
// reads the registry again. A registry that fails on that read refuses.
func (s *TargetPublicTestSuite) TestValidTargetReadsAgainOnAMiss() {
	tests := []struct {
		name   string
		second error
		want   bool
	}{
		{name: "a renamed host is found", want: true},
		{name: "a registry that fails refuses", second: errors.New("nats down"), want: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			calls := 0
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					calls++
					if calls == 1 {
						return []validation.AgentTarget{{MachineID: "m1", Hostname: "web-01"}}, nil
					}
					if tt.second != nil {
						return nil, tt.second
					}

					return []validation.AgentTarget{{MachineID: "m1", Hostname: "web-02"}}, nil
				},
			)

			_, ok := validation.Struct(context.Background(), targetInput{Target: "web-01"})
			s.True(ok, "fills the cache")
			_, ok = validation.Struct(context.Background(), targetInput{Target: "web-02"})
			s.Equal(tt.want, ok)
			s.Equal(2, calls)
		})
	}
}

func TestTargetPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TargetPublicTestSuite))
}

// TestResolveTargetEnforcing covers what changes once the controller verifies
// registrations: an unverified one stops deciding where work goes, and a
// contested hostname resolves the same way every time.
func (s *TargetPublicTestSuite) TestResolveTargetEnforcing() {
	tests := []struct {
		name         string
		enforce      bool
		agents       []validation.AgentTarget
		target       string
		validateFunc func(string)
	}{
		{
			name:    "a verified registration resolves to its machine ID",
			enforce: true,
			agents: []validation.AgentTarget{
				{MachineID: "machine-001", Hostname: "web-01", Verified: true},
			},
			target:       "web-01",
			validateFunc: func(got string) { s.Equal("machine-001", got) },
		},
		{
			name:    "an unverified registration does not claim the hostname",
			enforce: true,
			agents: []validation.AgentTarget{
				{MachineID: "machine-evil", Hostname: "web-01", Verified: false},
			},
			target: "web-01",
			validateFunc: func(got string) {
				// Unresolved, so the target is returned unchanged rather than
				// routed to a machine that never enrolled under this name.
				s.Equal("web-01", got)
			},
		},
		{
			name:    "a contested hostname resolves deterministically",
			enforce: true,
			agents: []validation.AgentTarget{
				{MachineID: "machine-zulu", Hostname: "web-01", Verified: true},
				{MachineID: "machine-alpha", Hostname: "web-01", Verified: true},
			},
			target: "web-01",
			validateFunc: func(got string) {
				s.Equal("machine-alpha", got)
			},
		},
		{
			name:    "with enforcement off an unverified registration still resolves",
			enforce: false,
			agents: []validation.AgentTarget{
				{MachineID: "machine-001", Hostname: "web-01", Verified: false},
			},
			target:       "web-01",
			validateFunc: func(got string) { s.Equal("machine-001", got) },
		},
		{
			name:    "a machine ID target is refused when unverified",
			enforce: true,
			agents: []validation.AgentTarget{
				{MachineID: "machine-001", Hostname: "web-01", Verified: false},
			},
			target:       "machine-001",
			validateFunc: func(got string) { s.Equal("machine-001", got) },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agents := tt.agents
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					return agents, nil
				},
			)
			validation.SetEnforceVerifiedRegistrations(tt.enforce)
			defer validation.SetEnforceVerifiedRegistrations(false)

			tt.validateFunc(validation.ResolveTarget(context.Background(), tt.target))
		})
	}
}

// TestValidTargetEnforcing covers the other two routing inputs: a hostname and
// a label are both refused when the registration behind them is unverified.
func (s *TargetPublicTestSuite) TestValidTargetEnforcing() {
	tests := []struct {
		name         string
		agents       []validation.AgentTarget
		target       string
		validateFunc func(bool)
	}{
		{
			name: "a verified hostname is a valid target",
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", Verified: true},
			},
			target:       "web-01",
			validateFunc: func(ok bool) { s.True(ok) },
		},
		{
			name: "an unverified hostname is not a valid target",
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", Verified: false},
			},
			target:       "web-01",
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "an unverified label is not a valid target",
			agents: []validation.AgentTarget{
				{
					MachineID: "m1",
					Hostname:  "web-01",
					Labels:    map[string]string{"group": "web"},
					Verified:  false,
				},
			},
			target:       "group:web",
			validateFunc: func(ok bool) { s.False(ok) },
		},
		{
			name: "a verified label is a valid target",
			agents: []validation.AgentTarget{
				{
					MachineID: "m1",
					Hostname:  "web-01",
					Labels:    map[string]string{"group": "web"},
					Verified:  true,
				},
			},
			target:       "group:web",
			validateFunc: func(ok bool) { s.True(ok) },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agents := tt.agents
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					return agents, nil
				},
			)
			validation.SetEnforceVerifiedRegistrations(true)
			defer validation.SetEnforceVerifiedRegistrations(false)

			_, ok := validation.Struct(context.Background(), targetInput{Target: tt.target})
			tt.validateFunc(ok)
		})
	}
}

// TestPendingTarget covers the hint that tells an operator to accept a pending
// agent, which is now derived from the target rather than from a variable two
// concurrent requests could overwrite.
func (s *TargetPublicTestSuite) TestPendingTarget() {
	tests := []struct {
		name         string
		agents       []validation.AgentTarget
		lister       bool
		target       string
		validateFunc func(string, bool)
	}{
		{
			name:   "a pending agent named by hostname",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Pending"},
			},
			target: "web-01",
			validateFunc: func(t string, ok bool) {
				s.True(ok)
				s.Equal("web-01", t)
			},
		},
		{
			name:   "a pending agent named by machine ID",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Pending"},
			},
			target: "m1",
			validateFunc: func(t string, ok bool) {
				s.True(ok)
				s.Equal("m1", t)
			},
		},
		{
			name:   "a ready agent is not pending",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Ready"},
			},
			target:       "web-01",
			validateFunc: func(_ string, ok bool) { s.False(ok) },
		},
		{
			name:   "a target matching no agent at all",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Pending"},
			},
			target:       "web-99",
			validateFunc: func(_ string, ok bool) { s.False(ok) },
		},
		{
			name:   "a broadcast when every agent is pending",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Pending"},
				{MachineID: "m2", Hostname: "web-02", State: "Pending"},
			},
			target: "_all",
			validateFunc: func(t string, ok bool) {
				s.True(ok)
				s.Equal("_all", t)
			},
		},
		{
			name:   "a broadcast when one agent is ready",
			lister: true,
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", State: "Pending"},
				{MachineID: "m2", Hostname: "web-02", State: "Ready"},
			},
			target:       "_any",
			validateFunc: func(_ string, ok bool) { s.False(ok) },
		},
		{
			name:         "no agents registered",
			lister:       true,
			agents:       []validation.AgentTarget{},
			target:       "web-01",
			validateFunc: func(_ string, ok bool) { s.False(ok) },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.lister {
				agents := tt.agents
				validation.RegisterTargetValidator(
					func(_ context.Context) ([]validation.AgentTarget, error) {
						return agents, nil
					},
				)
			}

			tt.validateFunc(validation.PendingTarget(context.Background(), tt.target))
		})
	}
}

// TestUnverifiedTarget covers the hint for a target whose agent is registered
// and that the controller does not route to: its registration is not signed
// by its admitted key, so "not found" would say the wrong thing
// (osapi-io/osapi#1484).
func (s *TargetPublicTestSuite) TestUnverifiedTarget() {
	tests := []struct {
		name     string
		agents   []validation.AgentTarget
		err      error
		target   string
		wantOK   bool
		contains []string
	}{
		{
			name:   "an unverified agent by hostname says why",
			agents: []validation.AgentTarget{{MachineID: "m1", Hostname: "debian-lts"}},
			target: "debian-lts",
			contains: []string{
				`agent "debian-lts" is registered`, "not signed by the key",
				"KeyMismatch", "osapi client agent get --hostname debian-lts",
			},
		},
		{
			name:     "an unverified agent by machine ID names its hostname",
			agents:   []validation.AgentTarget{{MachineID: "m1", Hostname: "debian-lts"}},
			target:   "m1",
			contains: []string{`agent "debian-lts" is registered`},
		},
		{
			name: "an unverified twin of a verified agent is no reason",
			agents: []validation.AgentTarget{
				{MachineID: "m1", Hostname: "web-01", Verified: true},
				{MachineID: "m2", Hostname: "web-01"},
			},
			target: "web-01",
			wantOK: true,
		},
		{
			name:     "an agent that does not exist is not found",
			agents:   []validation.AgentTarget{{MachineID: "m1", Hostname: "debian-lts"}},
			target:   "web-99",
			contains: []string{`target agent "web-99" not found`},
		},
		{
			name:     "a registry that cannot be read is not found",
			err:      errors.New("nats: timeout"),
			target:   "debian-lts",
			contains: []string{`target agent "debian-lts" not found`},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			agents, lerr := tt.agents, tt.err
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					return agents, lerr
				},
			)
			validation.SetEnforceVerifiedRegistrations(true)
			defer validation.SetEnforceVerifiedRegistrations(false)

			msg, ok := validation.Struct(context.Background(), targetInput{Target: tt.target})
			s.Equal(tt.wantOK, ok)
			for _, c := range tt.contains {
				s.Contains(msg, c)
			}
		})
	}
	s.Run("a verified agent with the name is no reason, called alone", func() {
		validation.RegisterTargetValidator(
			func(_ context.Context) ([]validation.AgentTarget, error) {
				return []validation.AgentTarget{
					{MachineID: "m2", Hostname: "web-01"},
					{MachineID: "m1", Hostname: "web-01", Verified: true},
				}, nil
			},
		)
		validation.SetEnforceVerifiedRegistrations(true)
		defer validation.SetEnforceVerifiedRegistrations(false)
		_, ok := validation.UnverifiedTarget(context.Background(), "web-01")
		s.False(ok)
	})
	s.Run("no lister", func() {
		validation.RegisterTargetValidator(nil)
		_, ok := validation.UnverifiedTarget(context.Background(), "web-01")
		s.False(ok)
	})
}
