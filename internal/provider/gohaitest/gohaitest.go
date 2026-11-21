// Copyright (c) 2024 John Dewey

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

// Package gohaitest builds a fake host for the tests of the providers
// that read through gohai: a filesystem with given files, and gohai's
// generated command runner, configured to answer given command lines.
package gohaitest

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/osapi-io/gohai/pkg/gohai/source"
	"github.com/osapi-io/gohai/pkg/gohai/source/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// FS is a memory filesystem that holds files, by path.
func FS(
	t testing.TB,
	files map[string]string,
) avfs.VFS {
	t.Helper()
	fs := memfs.New()
	for path, content := range files {
		require.NoError(t, fs.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, fs.WriteFile(path, []byte(content), 0o644))
	}

	return fs
}

// Commands is gohai's generated MockRunner that answers each command
// line, "name arg arg", with its output. Any other line fails, as a
// command the host does not have.
func Commands(
	t testing.TB,
	out map[string]string,
) *mocks.MockRunner {
	t.Helper()
	m := mocks.NewMockRunner(gomock.NewController(t))
	m.EXPECT().
		Execute(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, name string, args ...string) ([]byte, error) {
			line := strings.Join(append([]string{name}, args...), " ")
			if o, ok := out[line]; ok {
				return []byte(o), nil
			}

			return nil, errors.New("exec " + line + ": not found")
		}).
		AnyTimes()

	return m
}

// Linux is a Debian host with files and the output of commands.
func Linux(
	t testing.TB,
	files map[string]string,
	out map[string]string,
) source.Sources {
	return source.Sources{FS: FS(t, files), Exec: Commands(t, out), Family: "debian"}
}
