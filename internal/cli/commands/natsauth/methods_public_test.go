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

package natsauth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/natsauth"
	"github.com/osapi-io/osapi/internal/config"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) deps() natsauth.Deps {
	var cfg config.Config
	cfg.Controller.NATS.Namespace = "osapi"
	cfg.NATS.Stream = config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}

	return natsauth.Deps{FS: osfs.New(), Config: cfg}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	ctx := context.Background()

	s.Run("password: the files, and the server's block with --print", func() {
		out := filepath.Join(s.T().TempDir(), "nats")
		got, err := natsauth.Password(ctx, s.deps(), natsauth.CreateInput{Out: out, Print: true})
		s.Require().NoError(err)
		s.Len(got.Files, 2)
		s.Contains(got.Server, "authorization")
		info, err := os.Stat(out)
		s.Require().NoError(err)
		s.Equal("-rwx------", info.Mode().Perm().String(), "the directory is 0700")
	})

	s.Run("operator: no server part without --print", func() {
		got, err := natsauth.Operator(ctx, s.deps(), natsauth.CreateInput{Out: s.T().TempDir()})
		s.Require().NoError(err)
		s.Len(got.Files, 7)
		s.Empty(got.Server)
	})

	s.Run("a directory that cannot be made is the error", func() {
		file := filepath.Join(s.T().TempDir(), "file")
		s.Require().NoError(os.WriteFile(file, nil, 0o600))
		_, err := natsauth.Password(
			ctx,
			s.deps(),
			natsauth.CreateInput{Out: filepath.Join(file, "nats")},
		)
		s.ErrorContains(err, "make ")
	})

	s.Run("files that cannot be written are the error", func() {
		dir := s.T().TempDir()
		s.Require().NoError(os.MkdirAll(filepath.Join(dir, "osapi-nats.yaml", "x"), 0o700))
		_, err := natsauth.Password(ctx, s.deps(), natsauth.CreateInput{Out: dir})
		s.Error(err)
	})
}
