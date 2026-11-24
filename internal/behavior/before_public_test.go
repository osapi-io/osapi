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

package behavior_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/password"
	"github.com/osapi-io/osapi/internal/secret"
)

// pwIn is the input of a method whose controller step hashes a password:
// the caller sends password, and the agent gets only password_hash.
type pwIn struct {
	Name         string       `json:"name"                    validate:"required"  doc:"The user"     telemetry:"attr"`
	Password     secret.Value `json:"password,omitempty"      validate:"omitempty" doc:"The password"`
	PasswordHash string       `json:"password_hash,omitempty" validate:"omitempty" doc:"The hash"                      before:"set"`
}

func setPassword(
	_ context.Context,
	_ fakeProvider,
	in pwIn,
) (result, error) {
	return result{Key: in.Name, Changed: true}, nil
}

// hashStep is the step of the fixture: the password becomes its hash.
func hashStep(
	_ context.Context,
	in pwIn,
) (pwIn, error) {
	hash, err := password.Hash(in.Password.Reveal())
	if err != nil {
		return pwIn{}, err
	}
	in.PasswordHash, in.Password = hash, ""

	return in, nil
}

// withPassword is a served behavior with one write, its step step.
func withPassword(
	step any,
) (*behavior.Behavior, *behavior.Method) {
	b := behavior.Define("account", "Accounts", behavior.Node, onDebian,
		behavior.API, behavior.Area(area.Config),
		behavior.Write("update", setPassword).Describe("Update").Before(step),
	)

	return b, b.Method("update")
}

type BeforePublicTestSuite struct {
	suite.Suite
}

func (s *BeforePublicTestSuite) TestRunBefore() {
	tests := []struct {
		name         string
		step         any
		validateFunc func(out any, err error)
	}{
		{
			name: "when the step hashes the password, the job gets only the hash",
			step: hashStep,
			validateFunc: func(out any, err error) {
				s.Require().NoError(err)
				in := out.(pwIn)
				s.Equal("web", in.Name)
				s.Empty(in.Password, "the plain password is not in the job")
				s.True(strings.HasPrefix(in.PasswordHash, "$6$"))
			},
		},
		{
			name: "when the step fails, the job is not made",
			step: func(context.Context, pwIn) (pwIn, error) { return pwIn{}, errors.New("no salt") },
			validateFunc: func(out any, err error) {
				s.Nil(out)
				s.ErrorContains(err, "the controller step of account update: no salt")
			},
		},
		{
			name: "when the step leaves the secret set, the job gets it, for the controller to seal",
			step: func(_ context.Context, in pwIn) (pwIn, error) { return in, nil },
			validateFunc: func(out any, err error) {
				s.Require().NoError(err)
				s.Equal(secret.Value("hunter2"), out.(pwIn).Password)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b, m := withPassword(tc.step)
			s.True(m.HasBefore())
			tc.validateFunc(
				b.RunBefore(context.Background(), m, pwIn{Name: "web", Password: "hunter2"}),
			)
		})
	}

	s.Run("when the method has no step, the input is the job's", func() {
		b := top()
		m := b.Methods[0]
		s.False(m.HasBefore())
		out, err := b.RunBefore(context.Background(), m, getIn{Key: "k"})
		s.NoError(err)
		s.Equal(getIn{Key: "k"}, out)
	})
}

func (s *BeforePublicTestSuite) TestSetByBefore() {
	t := reflect.TypeFor[pwIn]()
	s.False(behavior.SetByBefore(t.Field(1)))
	s.True(behavior.SetByBefore(t.Field(2)))
}

func TestBeforePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BeforePublicTestSuite))
}
