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

package user_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/user"
)

type ReaderPublicTestSuite struct {
	suite.Suite
}

// TestWrites: off the Debian family the reader changes nothing. The
// definition limits the writes with On, so this is the reader's own answer.
func (s *ReaderPublicTestSuite) TestWrites() {
	p, err := user.NewReader(host.Host{})
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = p.Create(ctx, user.CreateInput{Name: "alice"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(ctx, user.UpdateInput{Name: "alice"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Delete(ctx, "alice")
	s.ErrorIs(err, behavior.ErrUnsupported)
	s.ErrorIs(p.SetPassword(ctx, "alice", "$6$x$y"), behavior.ErrUnsupported)
}

func (s *ReaderPublicTestSuite) TestIgnoreNotFound() {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "when the user is not found", err: fmt.Errorf("x: %w", behavior.ErrNotFound)},
		{name: "when the read failed", err: errors.New("boom"), want: errors.New("boom")},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, user.IgnoreNotFound(tc.err))
		})
	}
}

func (s *ReaderPublicTestSuite) TestSameGroups() {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "when the groups are the same in another order",
			a:    []string{"a", "b"},
			b:    []string{"b", "a"},
			want: true,
		},
		{name: "when one names a group twice", a: []string{"a"}, b: []string{"a", "a"}, want: true},
		{name: "when a group differs", a: []string{"a"}, b: []string{"b"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, user.SameGroups(tc.a, tc.b))
		})
	}
}

// TestSteps: each controller step leaves the job only the hash of the
// password, and a call with no password stays as it is.
func (s *ReaderPublicTestSuite) TestSteps() {
	ctx := context.Background()

	in, err := user.HashCreate(ctx, user.CreateInput{Name: "alice", Password: "pw"})
	s.Require().NoError(err)
	s.Empty(in.Password)
	s.True(strings.HasPrefix(in.PasswordHash, "$6$"))
	in, err = user.HashCreate(ctx, user.CreateInput{Name: "alice"})
	s.Require().NoError(err)
	s.Empty(in.PasswordHash, "no password, no hash")

	pw, err := user.HashPassword(ctx, user.PasswordInput{Name: "alice", Password: "pw"})
	s.Require().NoError(err)
	s.Empty(pw.Password)
	s.True(strings.HasPrefix(pw.PasswordHash, "$6$"))
	pw, err = user.HashPassword(ctx, user.PasswordInput{Name: "alice"})
	s.Require().NoError(err)
	s.Empty(pw.PasswordHash)

	defer user.SetHashFn(func(string) (string, error) { return "", errors.New("no salt") })()
	_, err = user.HashCreate(ctx, user.CreateInput{Name: "alice", Password: "pw"})
	s.ErrorContains(err, "no salt")
	_, err = user.HashPassword(ctx, user.PasswordInput{Name: "alice", Password: "pw"})
	s.ErrorContains(err, "no salt")
}

func TestReaderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReaderPublicTestSuite))
}
