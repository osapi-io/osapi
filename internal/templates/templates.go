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

// Package templates holds the system templates of osapi, which the controller
// keeps in the Object Store under osapi/.
package templates

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

//go:embed all:resources/templates
var systemTemplates embed.FS

// embeddedFS is a package-level variable for the walk filesystem.
// Swapped in internal tests to inject walk errors.
var embeddedFS fs.FS = systemTemplates

// readEmbeddedFile is a package-level variable for testing the read error path.
// Tests swap it via export_test.go.
var readEmbeddedFile = defaultReadEmbeddedFile

// defaultReadEmbeddedFile reads a template from the embedded filesystem.
func defaultReadEmbeddedFile(
	path string,
) ([]byte, error) {
	return systemTemplates.ReadFile(path)
}

// TemplateStore reads and writes the system templates in the Object Store
// bucket. A *natsclient.Client is one.
type TemplateStore interface {
	ObjectGetBytes(ctx context.Context, bucket string, name string) ([]byte, error)
	ObjectPutBytes(
		ctx context.Context,
		bucket string,
		name string,
		data []byte,
	) (*jetstream.ObjectInfo, error)
}

// SeedSystemTemplates uploads embedded osapi templates to the NATS
// object store with the "osapi/" prefix. Templates are updated if
// the embedded content differs from what's in the store (SHA comparison).
// These templates are protected from user deletion.
func SeedSystemTemplates(
	ctx context.Context,
	logger *slog.Logger,
	objStore TemplateStore,
	bucket string,
) error {
	logger = logger.With(slog.String("subsystem", "controller.templates"))

	var seeded int

	err := fs.WalkDir(embeddedFS, "resources/templates", func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Skip non-template files (e.g., .gitkeep placeholder required by go:embed).
		if d.Name() == ".gitkeep" {
			return nil
		}

		// Convert "resources/templates/osapi/cron-wrapper.tmpl" → "osapi/cron-wrapper.tmpl"
		objectName := path[len("resources/templates/"):]

		data, readErr := readEmbeddedFile(path)
		if readErr != nil {
			return fmt.Errorf("read embedded template %q: %w", path, readErr)
		}

		embeddedSHA := computeEmbeddedSHA256(data)

		// Check if the object already exists with the same content.
		existing, getErr := objStore.ObjectGetBytes(ctx, bucket, objectName)
		if getErr == nil {
			existingSHA := computeEmbeddedSHA256(existing)
			if existingSHA == embeddedSHA {
				logger.Debug(
					"osapi template unchanged, skipping",
					slog.String("name", objectName),
					slog.String("sha256", embeddedSHA),
				)
				return nil
			}

			logger.Info(
				"osapi template changed, updating",
				slog.String("name", objectName),
				slog.String("old_sha256", existingSHA),
				slog.String("new_sha256", embeddedSHA),
			)
		}

		if _, putErr := objStore.ObjectPutBytes(ctx, bucket, objectName, data); putErr != nil {
			return fmt.Errorf("upload osapi template %q: %w", objectName, putErr)
		}

		logger.Info(
			"seeded osapi template",
			slog.String("name", objectName),
			slog.String("sha256", embeddedSHA),
			slog.Int("size", len(data)),
		)

		seeded++

		return nil
	})
	if err != nil {
		return err
	}

	if seeded == 0 {
		logger.Debug("no osapi templates to seed")
	}

	return nil
}

// computeEmbeddedSHA256 returns the hex-encoded SHA-256 hash of data.
func computeEmbeddedSHA256(
	data []byte,
) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
