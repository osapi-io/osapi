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

package sshkey_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/sshkey"
)

type ReaderPublicTestSuite struct {
	suite.Suite
}

// TestWrites: off the Debian family the reader changes nothing. The
// definition limits the writes with On, so this is the reader's own answer.
func (s *ReaderPublicTestSuite) TestWrites() {
	p, err := sshkey.NewReader(host.Host{})
	s.Require().NoError(err)
	ctx := context.Background()

	_, _, err = p.Add(ctx, "alice", "ssh-ed25519 AAAA")
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(ctx, "alice", "SHA256:x")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *ReaderPublicTestSuite) TestFingerprint() {
	tests := []struct {
		name string
		line string
		want string
		err  string
	}{
		{
			name: "when the line has a key",
			line: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z c",
			want: "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
		},
		{
			name: "when options are before the key",
			line: `from="10.0.0.0/8" ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJl85dcLw7lIdX+IyXEXT00CTCezv/2E1jvolleleo6Z`,
			want: "SHA256:mnpIO1El26FsWnOeLzppADcNXgXRNHh9MPNPLPvSCwk",
		},
		{
			name: "when the line is a comment",
			line: "# ssh-ed25519 AAAA",
			err:  "not an SSH public key",
		},
		{name: "when the line is empty", line: "  ", err: "not an SSH public key"},
		{name: "when the key is not base64", line: "ssh-ed25519 !!!", err: "illegal base64"},
		{name: "when the line has no key type", line: "AAAA comment", err: "not an SSH public key"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := sshkey.Fingerprint(tc.line)
			if tc.err != "" {
				s.ErrorContains(err, tc.err)

				return
			}
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func TestReaderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReaderPublicTestSuite))
}
