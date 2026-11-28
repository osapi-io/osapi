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

// Package genfile writes generated files for the local, endpoint and
// behavior generators alike: it removes the files a run made before, writes
// each file under the repository's root, and groups a Go file's imports as
// the repository does.
//
// A Writer changes nothing on disk until Flush. Clean, Write, Go, Remove and
// Registry only record what to do, so a run that fails or panics part way,
// such as on a literal that jennifer cannot write, leaves the files of the
// last run as they were. Flush writes each file to a temporary file beside
// it and renames it, then removes the stale ones.
package genfile

import (
	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
)

// SetHostFS replaces the filesystem that the package reads and writes.
func SetHostFS(
	fs avfs.VFS,
) {
	hostFS = fs
}

// ResetHostFS puts back the filesystem of the host.
func ResetHostFS() {
	hostFS = osfs.NewWithNoIdm()
}
