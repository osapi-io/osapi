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

package controller

import (
	"context"
	"io"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	fileEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/file"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/templates"
)

// fileBuckets is the part of a NATS client that the file endpoints use.
type fileBuckets interface {
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
	KVKeys(ctx context.Context, bucket string, filters ...string) ([]string, error)
	KVGet(ctx context.Context, bucket string, key string) (jetstream.KeyValueEntry, error)
	ObjectStoreCreateOrUpdate(ctx context.Context, config jetstream.ObjectStoreConfig) error
	ObjectPutBytes(
		ctx context.Context,
		bucket string,
		name string,
		data []byte,
	) (*jetstream.ObjectInfo, error)
	KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
}

// seedTemplates puts osapi's system templates in the Object Store. Tests
// swap it.
var seedTemplates = templates.SeedSystemTemplates

// openFileBuckets opens the Object Store of the file endpoints and the
// file-state bucket of the stale deployment checks, and returns their names.
// Each one is "" if cfg names none, or if it does not open. The controller runs without the file
// endpoints then, so a failure is a warning.
func openFileBuckets(
	ctx context.Context,
	log *slog.Logger,
	nc fileBuckets,
	namespace string,
	cfg config.NATS,
) (objBucket string, stateBucket string) {
	if cfg.Objects.Bucket != "" {
		objCfg := bus.ObjectStore(namespace, cfg.Objects)
		if err := nc.ObjectStoreCreateOrUpdate(ctx, objCfg); err != nil {
			log.Warn("Object Store not available, file endpoints disabled",
				slog.String("bucket", objCfg.Bucket),
				slog.String("error", err.Error()))
		} else {
			objBucket = objCfg.Bucket
			if err := seedTemplates(ctx, log, nc, objBucket); err != nil {
				log.Warn("failed to seed system templates",
					slog.String("error", err.Error()))
			}
		}
	}

	if cfg.FileState.Bucket != "" {
		name := job.ApplyNamespaceToInfraName(namespace, cfg.FileState.Bucket)
		if _, err := nc.KVStatus(ctx, name); err != nil {
			log.Warn("file state KV not available",
				slog.String("error", err.Error()))
		} else {
			stateBucket = name
		}
	}

	return objBucket, stateBucket
}

// fileDeps is what the file group uses: the Object Store and the file-state
// bucket, each bound to its name, and nil when it is not open.
func fileDeps(
	nc fileBuckets,
	objBucket string,
	stateBucket string,
) fileEndpoint.Deps {
	var d fileEndpoint.Deps
	if objBucket != "" {
		d.Objects = objects{nc: nc, bucket: objBucket}
	}
	if stateBucket != "" {
		d.States = states{nc: nc, bucket: stateBucket}
	}

	return d
}

// objects is the Object Store bucket of the files.
type objects struct {
	nc     fileBuckets
	bucket string
}

func (o objects) Put(
	ctx context.Context,
	meta jetstream.ObjectMeta,
	r io.Reader,
) (*jetstream.ObjectInfo, error) {
	return o.nc.ObjectPut(ctx, o.bucket, meta, r)
}

func (o objects) GetBytes(
	ctx context.Context,
	name string,
) ([]byte, error) {
	return o.nc.ObjectGetBytes(ctx, o.bucket, name)
}

func (o objects) Info(
	ctx context.Context,
	name string,
) (*jetstream.ObjectInfo, error) {
	return o.nc.ObjectInfo(ctx, o.bucket, name)
}

func (o objects) Delete(
	ctx context.Context,
	name string,
) error {
	return o.nc.ObjectDelete(ctx, o.bucket, name)
}

func (o objects) List(
	ctx context.Context,
) ([]*jetstream.ObjectInfo, error) {
	return o.nc.ObjectList(ctx, o.bucket)
}

// states is the KV bucket of the deploy state.
type states struct {
	nc     fileBuckets
	bucket string
}

func (s states) Keys(
	ctx context.Context,
) ([]string, error) {
	return s.nc.KVKeys(ctx, s.bucket)
}

func (s states) Get(
	ctx context.Context,
	key string,
) (jetstream.KeyValueEntry, error) {
	return s.nc.KVGet(ctx, s.bucket, key)
}
