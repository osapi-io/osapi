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

package agent

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

//go:embed templates/*
var systemTemplates embed.FS

// SeedSystemTemplates uploads embedded system templates to the NATS
// object store with the "system/" prefix. Existing objects are skipped
// (idempotent). These templates are protected from user deletion.
func SeedSystemTemplates(
	ctx context.Context,
	logger *slog.Logger,
	objStore jetstream.ObjectStore,
) error {
	return fs.WalkDir(systemTemplates, "templates", func(
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

		// Convert "templates/system/cron-wrapper.tmpl" → "system/cron-wrapper.tmpl"
		objectName := path[len("templates/"):]

		// Check if already exists (idempotent).
		if _, infoErr := objStore.GetInfo(ctx, objectName); infoErr == nil {
			logger.Debug(
				"system template already exists, skipping",
				slog.String("name", objectName),
			)
			return nil
		}

		data, readErr := systemTemplates.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read embedded template %q: %w", path, readErr)
		}

		if _, putErr := objStore.PutBytes(ctx, objectName, data); putErr != nil {
			return fmt.Errorf("upload system template %q: %w", objectName, putErr)
		}

		logger.Info(
			"seeded system template",
			slog.String("name", objectName),
			slog.Int("size", len(data)),
		)

		return nil
	})
}
