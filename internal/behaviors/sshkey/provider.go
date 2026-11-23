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

package sshkey

import "context"

// Provider reads and changes the authorized SSH keys of a user.
type Provider interface {
	// Keys is the keys of the user's authorized_keys. A user that the host
	// does not have is not found; one with no file has none.
	Keys(ctx context.Context, name string) ([]Key, error)
	// Add appends the key line to the user's authorized_keys, and returns
	// its fingerprint. A key that the file has already is unchanged.
	Add(ctx context.Context, name string, line string) (string, bool, error)
	// Remove drops each line of the key with the fingerprint.
	Remove(ctx context.Context, name string, fingerprint string) (bool, error)
}
