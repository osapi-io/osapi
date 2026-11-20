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

package controllerkey_test

import (
	"context"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/rofs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/controllerkey"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestFingerprint() {
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/pki", 0o700))
	d := controllerkey.Deps{FS: fs, KeyDir: "/pki"}

	first, err := controllerkey.Fingerprint(
		context.Background(),
		d,
		controllerkey.FingerprintInput{},
	)
	s.Require().NoError(err)
	s.NotEmpty(first.Fingerprint, "a key pair is made when there is none")
	s.Equal("/pki", first.KeyDir)

	again, err := controllerkey.Fingerprint(
		context.Background(),
		d,
		controllerkey.FingerprintInput{},
	)
	s.Require().NoError(err)
	s.Equal(first.Fingerprint, again.Fingerprint, "the second run loads the same key")

	_, err = controllerkey.Fingerprint(
		context.Background(),
		controllerkey.Deps{
			FS:     rofs.New(memfs.New()),
			KeyDir: "/pki",
		},
		controllerkey.FingerprintInput{},
	)
	s.ErrorContains(err, "load the controller key from /pki")
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
