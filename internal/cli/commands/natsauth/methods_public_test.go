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

	"github.com/osapi-io/osapi/internal/audit"

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

// TestCreateAudited: making credentials is a local change in the audit
// log, with the type and the count of files, and no secret
// (osapi-io/osapi#1453).
func (s *MethodsPublicTestSuite) TestCreateAudited() {
	var entries []audit.Entry
	d := s.deps()
	d.Audit = func(_ context.Context, e audit.Entry) error {
		entries = append(entries, e)

		return nil
	}
	out := s.T().TempDir()
	_, err := natsauth.Operator(context.Background(), d, natsauth.CreateInput{Out: out})
	s.Require().NoError(err)
	s.Require().Len(entries, 1)
	s.Equal(audit.ActorLocal, entries[0].Actor)
	s.Equal("natsauth.create", entries[0].Action)
	s.Equal(out, entries[0].Target)
	s.Equal("type=operator files=10", entries[0].RequestSummary)
}

// TestOperatorUpdate: an update signs the files of an operator setup again,
// with the same keys, and is audited as an update.
func (s *MethodsPublicTestSuite) TestOperatorUpdate() {
	ctx := context.Background()
	s.Run("the files of operator create, again, with --print", func() {
		var entries []audit.Entry
		d := s.deps()
		d.Audit = func(_ context.Context, e audit.Entry) error {
			entries = append(entries, e)

			return nil
		}
		out := s.T().TempDir()
		_, err := natsauth.Operator(ctx, d, natsauth.CreateInput{Out: out})
		s.Require().NoError(err)
		seed, _ := os.ReadFile(filepath.Join(out, "account.nk"))

		got, err := natsauth.OperatorUpdate(ctx, d, natsauth.UpdateInput{Out: out, Print: true})
		s.Require().NoError(err)
		s.Len(got.Files, 10)
		s.Contains(got.Server, "resolver_preload")
		after, _ := os.ReadFile(filepath.Join(out, "account.nk"))
		s.Equal(seed, after, "the account keeps its key")
		s.Require().Len(entries, 2)
		s.Equal("natsauth.update", entries[1].Action)
		s.Equal("type=operator files=10", entries[1].RequestSummary)
	})
	s.Run("a directory with no operator setup", func() {
		_, err := natsauth.OperatorUpdate(ctx, s.deps(), natsauth.UpdateInput{Out: s.T().TempDir()})
		s.ErrorContains(err, "read operator.nk")
	})
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
		s.Len(got.Files, 10)
		s.Empty(got.Server)
	})

	s.Run("nkey: a seed for each role and the server's block", func() {
		got, err := natsauth.NKey(
			ctx,
			s.deps(),
			natsauth.CreateInput{Out: s.T().TempDir(), Print: true},
		)
		s.Require().NoError(err)
		s.Len(got.Files, 5)
		s.Contains(got.Server, "nkey:")
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
