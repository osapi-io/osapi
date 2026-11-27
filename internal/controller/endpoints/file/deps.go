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

	"github.com/nats-io/nats.go/jetstream"
)

// Deps is what the file group's handlers use.
type Deps struct {
	// Objects is the Object Store bucket of the files. Nil when the
	// controller has no object bucket; each method then answers 503.
	Objects Objects
	// States is the KV bucket of the deploy state. Nil when it is not set;
	// stale then answers 503.
	States States
}

// Objects is the Object Store bucket of the files.
type Objects interface {
	Put(ctx context.Context, meta jetstream.ObjectMeta, r io.Reader) (*jetstream.ObjectInfo, error)
	Info(ctx context.Context, name string) (*jetstream.ObjectInfo, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context) ([]*jetstream.ObjectInfo, error)
}

// States is the KV bucket of the deploy state: one entry for each file
// that an agent deployed.
type States interface {
	Keys(ctx context.Context) ([]string, error)
	Get(ctx context.Context, key string) (jetstream.KeyValueEntry, error)
}
