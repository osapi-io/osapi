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
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api/file"
	"github.com/osapi-io/osapi/internal/job"
)

// fileBuckets is the part of a NATS client that the file endpoints use.
type fileBuckets interface {
	file.ObjectStoreManager
	file.StateKeyValue
	ObjectStoreCreateOrUpdate(ctx context.Context, config jetstream.ObjectStoreConfig) error
	KVStatus(ctx context.Context, bucket string) (jetstream.KeyValueStatus, error)
}

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
