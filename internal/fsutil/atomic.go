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

// Package fsutil holds filesystem operations shared by providers that write
// files a running system depends on.
package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"sync/atomic"

	"github.com/avfs/avfs"
)

// tmpCounter distinguishes two temporary files written by the same process at
// the same instant. Together with the process ID it makes the name unique
// without needing the filesystem to generate one.
var tmpCounter atomic.Uint64

// WriteFileAtomic writes content to path so that a reader never sees a partial
// file: the content goes to a temporary file beside the target, is given the
// requested mode, and is then renamed over it. A rename within one directory is
// atomic, so the target is either the old file or the new one, never half of
// either.
//
// The temporary file sits in the target's own directory because a rename across
// filesystems is not atomic, and /tmp is frequently its own filesystem. It is
// removed on every failure path, so a failed write leaves nothing behind.
//
// The mode is set on the temporary file rather than on the target, so the content
// is never briefly readable by someone the mode excludes. It is set twice — at
// creation and again explicitly — because the mode passed to open is masked by
// the process umask and chmod is not.
func WriteFileAtomic(
	vfs avfs.VFS,
	path string,
	content []byte,
	mode fs.FileMode,
) error {
	tmpPath := fmt.Sprintf(
		"%s.osapi-%d-%d",
		path,
		os.Getpid(),
		tmpCounter.Add(1),
	)

	tmp, err := vfs.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create temp file %q: %w", tmpPath, err)
	}

	cleanup := func() {
		_ = tmp.Close()
		_ = vfs.Remove(tmpPath)
	}

	if _, err := tmp.Write(content); err != nil {
		cleanup()

		return fmt.Errorf("write temp file %q: %w", tmpPath, err)
	}

	if err := vfs.Chmod(tmpPath, mode); err != nil {
		cleanup()

		return fmt.Errorf("set mode on %q: %w", tmpPath, err)
	}

	if err := tmp.Close(); err != nil {
		_ = vfs.Remove(tmpPath)

		return fmt.Errorf("close temp file %q: %w", tmpPath, err)
	}

	if err := vfs.Rename(tmpPath, path); err != nil {
		_ = vfs.Remove(tmpPath)

		return fmt.Errorf("rename %q to %q: %w", tmpPath, path, err)
	}

	return nil
}
