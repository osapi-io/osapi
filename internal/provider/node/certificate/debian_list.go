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

package certificate

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// List returns all system and osapi-managed CA certificates.
func (d *Debian) List(
	ctx context.Context,
) ([]Entry, error) {
	systemCAs, err := d.listSystemCAs()
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}

	customCAs := d.listCustomCAs(ctx)

	result := make([]Entry, 0, len(customCAs)+len(systemCAs))
	result = append(result, customCAs...)
	result = append(result, systemCAs...)

	return result, nil
}

// listSystemCAs walks /usr/share/ca-certificates/ recursively, returning
// entries with Source "system". The .crt suffix is stripped and the
// directory prefix is removed to produce clean names.
func (d *Debian) listSystemCAs() ([]Entry, error) {
	var entries []Entry

	err := d.fs.WalkDir(systemCADir, func(
		path string,
		dirEntry fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".crt") {
			return nil
		}

		// Strip the directory prefix and .crt suffix.
		name := strings.TrimPrefix(path, systemCADir+"/")
		name = strings.TrimSuffix(name, ".crt")

		entries = append(entries, Entry{
			Name:   name,
			Source: "system",
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// listCustomCAs reads /usr/local/share/ca-certificates/ for managed
// certificates. Only files with the osapi- prefix and .crt suffix that
// are tracked in the file-state KV (and not undeployed) are included.
func (d *Debian) listCustomCAs(
	ctx context.Context,
) []Entry {
	var entries []Entry

	dirEntries, err := d.fs.ReadDir(customCADir)
	if err != nil {
		return entries
	}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		fileName := dirEntry.Name()

		if !strings.HasPrefix(fileName, managedPrefix) {
			continue
		}

		if !strings.HasSuffix(fileName, ".crt") {
			continue
		}

		path := filepath.Join(customCADir, fileName)
		if !d.isManagedFile(ctx, path) {
			continue
		}

		// Strip osapi- prefix and .crt suffix for clean name.
		name := strings.TrimPrefix(fileName, managedPrefix)
		name = strings.TrimSuffix(name, ".crt")

		entries = append(entries, Entry{
			Name:   name,
			Source: "custom",
		})
	}

	return entries
}
