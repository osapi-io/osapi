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

package controller_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"

	"github.com/avfs/avfs"
)

type VaultsPublicTestSuite struct {
	suite.Suite
}

func TestVaultsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(VaultsPublicTestSuite))
}

func (s *VaultsPublicTestSuite) TestMakeVaultKeys() {
	ctx := context.Background()
	vfs := osfs.NewWithNoIdm()

	s.Run("when the key file is not there, it makes it with mode 0600 and logs it", func() {
		file := filepath.Join(s.T().TempDir(), "osapi", "secrets.key")
		var log bytes.Buffer
		err := controller.ExportMakeVaultKeys(ctx, slog.New(slog.NewTextHandler(&log, nil)), vfs,
			[]vault.Spec{{Name: "default", Type: "nats", KeyFile: file}})
		s.Require().NoError(err)
		info, err := testFS.Stat(file)
		s.Require().NoError(err)
		s.Equal(os.FileMode(0o600), info.Mode().Perm())
		s.Contains(log.String(), "vault key created")
		s.Contains(log.String(), file)
	})

	s.Run("when the key file is there, it never replaces it", func() {
		file := filepath.Join(s.T().TempDir(), "secrets.key")
		ring, err := natsvault.CreateKey(ctx, vfs, file)
		s.Require().NoError(err)
		before, _ := avfs.ReadFile(testFS, file)
		var log bytes.Buffer
		s.Require().NoError(controller.ExportMakeVaultKeys(ctx,
			slog.New(slog.NewTextHandler(&log, nil)), vfs,
			[]vault.Spec{{Name: "default", Type: "nats", KeyFile: file}}))
		after, _ := avfs.ReadFile(testFS, file)
		s.Equal(before, after, "the key "+ring.ID()+" stays")
		s.Empty(log.String())
	})

	s.Run(
		"when a vault has its key in the environment, of another type or no file, it makes none",
		func() {
			dir := s.T().TempDir()
			s.Require().
				NoError(controller.ExportMakeVaultKeys(ctx, slog.Default(), vfs, []vault.Spec{
					{
						Name:    "env",
						Type:    "nats",
						KeyFile: filepath.Join(dir, "env.key"),
						KeyEnv:  "KEY",
					},
					{Name: "other", Type: "vault", KeyFile: filepath.Join(dir, "other.key")},
					{Name: "none", Type: "nats"},
				}))
			entries, _ := testFS.ReadDir(dir)
			s.Empty(entries)
		},
	)
}
