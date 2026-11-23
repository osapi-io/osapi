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

package file

import (
	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
)

// NewLinux is the provider on the building blocks' providers, for the
// provider's tests, with facts when facts is not nil.
func NewLinux(
	objects object.Provider,
	files managedfile.Provider,
	fs avfs.VFS,
	hostname string,
	facts map[string]any,
) Provider {
	l := &linux{objects: objects, files: files, fs: fs, hostname: hostname}
	if facts != nil {
		l.facts = func() map[string]any { return facts }
	}

	return l
}
