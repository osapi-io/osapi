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

package object

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// store reads the objects through the host's object store. It is the same on
// each platform: the store is NATS, not the host.
type store struct {
	objects host.Objects
}

// newStore is the provider of each platform. An agent with no object store
// cannot read an object, so the behaviors that require this one report
// skipped there.
func newStore(
	h host.Host,
) (Provider, error) {
	if h.Objects == nil {
		return nil, behavior.ErrUnsupported
	}

	return &store{objects: h.Objects}, nil
}

// Object is the object name, with the SHA-256 of its content.
func (s *store) Object(
	ctx context.Context,
	name string,
) (Object, error) {
	content, contentType, err := s.objects.Object(ctx, name)
	if err != nil {
		return Object{}, err
	}
	sum := sha256.Sum256(content)

	return Object{
		Name:        name,
		Content:     string(content),
		ContentType: contentType,
		SHA256:      hex.EncodeToString(sum[:]),
	}, nil
}
