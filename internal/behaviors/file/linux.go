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
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/behaviors/object"
	"github.com/osapi-io/osapi/internal/provider"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// linux deploys objects through the object and managedfile building blocks.
// The agent gives it its facts, which a template reads as .Facts.
type linux struct {
	provider.FactsAware

	objects  object.Provider
	files    managedfile.Provider
	fs       avfs.VFS
	hostname string
}

// newLinux is the provider on each Linux host. It uses the providers of the
// behaviors it requires.
func newLinux(
	h host.Host,
) (Provider, error) {
	return &linux{
		objects:  host.Require[object.Provider](h, object.Behavior),
		files:    host.Require[managedfile.Provider](h, managedfile.Behavior),
		fs:       h.FS,
		hostname: h.Hostname,
	}, nil
}

// Deploy reads the object, renders it when it is a template, and writes it
// through managedfile, which records it.
func (l *linux) Deploy(
	ctx context.Context,
	in UpdateInput,
) (UpdateResult, error) {
	obj, err := behavior.Call(ctx, object.Behavior, object.Get, l.objects,
		object.GetInput{Name: in.ObjectName})
	if err != nil {
		return UpdateResult{}, err
	}

	contentType := in.ContentType
	if contentType == "" {
		contentType = obj.ContentType
	}
	if contentType == "" {
		contentType = string(client.ContentRaw)
	}
	content := obj.Content
	if contentType == string(client.ContentTemplate) {
		content, err = behavior.Render(content, behavior.TemplateData{
			Facts:    l.Facts(),
			Vars:     in.Vars,
			Hostname: l.hostname,
		})
		if err != nil {
			return UpdateResult{}, err
		}
	}

	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, l.files,
		managedfile.UpdateInput{
			Path:        in.Path,
			Content:     content,
			Mode:        in.Mode,
			Owner:       in.Owner,
			Group:       in.Group,
			Object:      in.ObjectName,
			ContentType: contentType,
		})
	if err != nil {
		return UpdateResult{}, err
	}

	return UpdateResult{
		Path:    in.Path,
		SHA256:  checksum([]byte(content)),
		Changed: result.Changed,
	}, nil
}

// Undeploy removes the file through managedfile, which keeps its record.
func (l *linux) Undeploy(
	ctx context.Context,
	path string,
) (bool, error) {
	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Delete, l.files,
		managedfile.DeleteInput{Path: path})
	if err != nil {
		return false, err
	}

	return result.Changed, nil
}

// Status compares the file on the host with the record that managedfile
// keeps of what osapi wrote.
func (l *linux) Status(
	ctx context.Context,
	path string,
) (Status, error) {
	record, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Get, l.files,
		managedfile.GetInput{Path: path})
	switch {
	case errors.Is(err, managedfile.ErrNotManaged):
		return Status{Path: path, State: Missing}, nil
	case err != nil:
		return Status{}, err
	case record.Undeployed:
		return Status{Path: path, State: Missing}, nil
	}

	data, err := l.fs.ReadFile(path)
	if err != nil {
		return Status{Path: path, State: Missing}, nil
	}
	sum := checksum(data)
	if sum == record.SHA256 {
		return Status{Path: path, State: InSync, SHA256: sum}, nil
	}

	return Status{Path: path, State: Drifted, SHA256: sum}, nil
}

// checksum is the SHA-256 of data, in hex.
func checksum(
	data []byte,
) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}
