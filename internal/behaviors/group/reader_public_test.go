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

package group_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/group"
)

type ReaderPublicTestSuite struct {
	suite.Suite
}

// TestWrites: off the Debian family the reader changes nothing. The
// definition limits the writes with On, so this is the reader's own answer.
func (s *ReaderPublicTestSuite) TestWrites() {
	p, err := group.NewReader(host.Host{})
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = p.Create(ctx, group.CreateInput{Name: "dev"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(ctx, "dev", nil)
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Delete(ctx, "dev")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *ReaderPublicTestSuite) TestIgnoreNotFound() {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "when the group is not found", err: fmt.Errorf("x: %w", behavior.ErrNotFound)},
		{name: "when the read failed", err: errors.New("boom"), want: errors.New("boom")},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, group.IgnoreNotFound(tc.err))
		})
	}
}

func (s *ReaderPublicTestSuite) TestSameMembers() {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "when the members are the same in another order",
			a:    []string{"alice", "bob"},
			b:    []string{"bob", "alice"},
			want: true,
		},
		{
			name: "when one names a member twice",
			a:    []string{"alice"},
			b:    []string{"alice", "alice"},
			want: true,
		},
		{name: "when both have none", want: true},
		{name: "when a member differs", a: []string{"alice"}, b: []string{"bob"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, group.SameMembers(tc.a, tc.b))
		})
	}
}

func TestReaderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReaderPublicTestSuite))
}
