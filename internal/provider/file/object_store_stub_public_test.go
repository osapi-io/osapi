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

package file_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/nats-io/nats.go/jetstream"
)

// computeTestSHA256 returns the hex-encoded SHA-256 hash of the given data.
func computeTestSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

// stubObjectStore is a minimal test stub implementing jetstream.ObjectStore.
// Only GetBytes is functional; all other methods panic if called.
type stubObjectStore struct {
	getBytesData []byte
	getBytesErr  error
}

func (s *stubObjectStore) GetBytes(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) ([]byte, error) {
	return s.getBytesData, s.getBytesErr
}

func (s *stubObjectStore) Put(
	_ context.Context,
	_ jetstream.ObjectMeta,
	_ io.Reader,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: Put not implemented")
}

func (s *stubObjectStore) PutBytes(
	_ context.Context,
	_ string,
	_ []byte,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: PutBytes not implemented")
}

func (s *stubObjectStore) PutString(
	_ context.Context,
	_ string,
	_ string,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: PutString not implemented")
}

func (s *stubObjectStore) PutFile(
	_ context.Context,
	_ string,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: PutFile not implemented")
}

func (s *stubObjectStore) Get(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) (jetstream.ObjectResult, error) {
	panic("stubObjectStore: Get not implemented")
}

func (s *stubObjectStore) GetString(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) (string, error) {
	panic("stubObjectStore: GetString not implemented")
}

func (s *stubObjectStore) GetFile(
	_ context.Context,
	_ string,
	_ string,
	_ ...jetstream.GetObjectOpt,
) error {
	panic("stubObjectStore: GetFile not implemented")
}

func (s *stubObjectStore) GetInfo(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectInfoOpt,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: GetInfo not implemented")
}

func (s *stubObjectStore) UpdateMeta(
	_ context.Context,
	_ string,
	_ jetstream.ObjectMeta,
) error {
	panic("stubObjectStore: UpdateMeta not implemented")
}

func (s *stubObjectStore) Delete(
	_ context.Context,
	_ string,
) error {
	panic("stubObjectStore: Delete not implemented")
}

func (s *stubObjectStore) AddLink(
	_ context.Context,
	_ string,
	_ *jetstream.ObjectInfo,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: AddLink not implemented")
}

func (s *stubObjectStore) AddBucketLink(
	_ context.Context,
	_ string,
	_ jetstream.ObjectStore,
) (*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: AddBucketLink not implemented")
}

func (s *stubObjectStore) Seal(
	_ context.Context,
) error {
	panic("stubObjectStore: Seal not implemented")
}

func (s *stubObjectStore) Watch(
	_ context.Context,
	_ ...jetstream.WatchOpt,
) (jetstream.ObjectWatcher, error) {
	panic("stubObjectStore: Watch not implemented")
}

func (s *stubObjectStore) List(
	_ context.Context,
	_ ...jetstream.ListObjectsOpt,
) ([]*jetstream.ObjectInfo, error) {
	panic("stubObjectStore: List not implemented")
}

func (s *stubObjectStore) Status(
	_ context.Context,
) (jetstream.ObjectStoreStatus, error) {
	panic("stubObjectStore: Status not implemented")
}
