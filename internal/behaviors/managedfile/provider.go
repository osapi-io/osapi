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

package managedfile

import (
	"context"
	"errors"
)

// The causes Read reports that a caller answers differently.
var (
	// ErrNotManaged means osapi has no record of the file.
	ErrNotManaged = errors.New("not managed")
	// ErrUnreadable means the file's record is there but cannot be read.
	ErrUnreadable = errors.New("record unreadable")
)

// Provider is what each platform implements.
type Provider interface {
	// Write makes f.Path hold f.Content with f's mode and ownership, records
	// it, and reports whether anything changed. A file already holding the
	// content with that mode and owner is left alone, whatever its record
	// says.
	Write(ctx context.Context, f File) (bool, error)
	// Remove deletes path if osapi manages it, records that it no longer
	// does, and reports whether a file was deleted.
	Remove(ctx context.Context, path string) (bool, error)
	// Read is the record of path: ErrNotManaged without one, ErrUnreadable
	// when it cannot be read.
	Read(ctx context.Context, path string) (Record, error)
}
