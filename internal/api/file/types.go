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
	"context"
	"io"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

// ObjectStoreManager wraps the subset of jetstream.ObjectStore methods
// needed by the file API handlers. This minimal interface enables
// straightforward mocking in tests.
type ObjectStoreManager interface {
	// Put stores data from a reader under the given metadata.
	Put(
		ctx context.Context,
		meta jetstream.ObjectMeta,
		reader io.Reader,
	) (*jetstream.ObjectInfo, error)

	// PutBytes stores data under the given name.
	PutBytes(ctx context.Context, name string, data []byte) (*jetstream.ObjectInfo, error)

	// GetBytes retrieves the content stored under the given name.
	GetBytes(ctx context.Context, name string, opts ...jetstream.GetObjectOpt) ([]byte, error)

	// GetInfo retrieves metadata for the named object.
	GetInfo(
		ctx context.Context,
		name string,
		opts ...jetstream.GetObjectInfoOpt,
	) (*jetstream.ObjectInfo, error)

	// Delete removes the named object.
	Delete(ctx context.Context, name string) error

	// List returns metadata for all objects in the store.
	List(ctx context.Context, opts ...jetstream.ListObjectsOpt) ([]*jetstream.ObjectInfo, error)
}

// File implementation of the File APIs operations.
type File struct {
	objStore ObjectStoreManager
	logger   *slog.Logger
}
