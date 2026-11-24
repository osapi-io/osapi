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

package agentkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/rofs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/cli/commands/agentkey"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestFingerprint() {
	web := func() (*identity.Identity, error) {
		return &identity.Identity{MachineID: "m-1", Hostname: "web-01"}, nil
	}
	writable := func() avfs.VFS {
		fs := memfs.New()
		s.Require().NoError(fs.MkdirAll("/pki", 0o700))

		return fs
	}

	tests := []struct {
		name         string
		fs           avfs.VFS
		identity     func() (*identity.Identity, error)
		validateFunc func(agentkey.FingerprintOutput, error)
	}{
		{
			name:     "when the key and the identity are read, it is what admits the machine",
			fs:       writable(),
			identity: web,
			validateFunc: func(got agentkey.FingerprintOutput, err error) {
				s.Require().NoError(err)
				s.Equal("m-1", got.MachineID)
				s.Equal("web-01", got.Hostname)
				s.NotEmpty(got.Fingerprint)
				s.Equal(
					"osapi client admission admit --machine-id m-1 --fingerprint "+got.Fingerprint,
					got.Admit,
				)
			},
		},
		{
			name:     "when the key cannot be made, it is the error",
			fs:       rofs.New(memfs.New()),
			identity: web,
			validateFunc: func(_ agentkey.FingerprintOutput, err error) {
				s.ErrorContains(err, "load the agent key from /pki")
			},
		},
		{
			name: "when the identity cannot be read, it is the error",
			fs:   writable(),
			identity: func() (*identity.Identity, error) {
				return nil, errors.New("resolve machine-id: none")
			},
			validateFunc: func(_ agentkey.FingerprintOutput, err error) {
				s.EqualError(err, "resolve machine-id: none")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := agentkey.Deps{FS: tc.fs, KeyDir: "/pki", Identity: tc.identity}
			tc.validateFunc(
				agentkey.Fingerprint(context.Background(), d, agentkey.FingerprintInput{}),
			)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
