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

// StateKeyValue reads the file-state KV bucket, for the stale check. A
// *natsclient.Client is one.
type StateKeyValue interface {
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
}

// ObjectStoreManager is the access to the Object Store bucket that the file
// API handlers use. A *natsclient.Client is one.
type ObjectStoreManager interface {
	ObjectPut(
		ctx context.Context,
		bucket string,
		meta jetstream.ObjectMeta,
		reader io.Reader,
	) (*jetstream.ObjectInfo, error)
	ObjectGetBytes(ctx context.Context, bucket string, name string) ([]byte, error)
	ObjectInfo(ctx context.Context, bucket string, name string) (*jetstream.ObjectInfo, error)
	ObjectDelete(ctx context.Context, bucket string, name string) error
	ObjectList(ctx context.Context, bucket string) ([]*jetstream.ObjectInfo, error)
}

// File implementation of the File APIs operations.
type File struct {
	objStore    ObjectStoreManager
	objBucket   string
	stateKV     StateKeyValue
	stateBucket string
	logger      *slog.Logger
}
