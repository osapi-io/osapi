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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// contentTypeHeader is the header of an object that holds how a deploy
// treats it.
const contentTypeHeader = "Osapi-Content-Type"

// List is each stored file.
func List(
	ctx context.Context,
	d Deps,
	_ ListInput,
) (FileListResponse, error) {
	if d.Objects == nil {
		return FileListResponse{}, noStore()
	}
	objects, err := d.Objects.List(ctx)
	if err != nil {
		return FileListResponse{}, endpoint.Internal("failed to list files", err)
	}

	files := make([]FileInfo, 0, len(objects))
	for _, obj := range objects {
		if obj.Deleted {
			continue
		}
		source := "user"
		if protected(obj.Name) {
			source = "osapi"
		}
		files = append(files, FileInfo{
			Name:        obj.Name,
			Sha256:      hexDigest(obj.Digest),
			Size:        int(obj.Size), //nolint:gosec // an object's size fits an int
			ContentType: contentType(obj),
			Source:      source,
		})
	}

	return FileListResponse{Items: files, TotalItems: len(files)}, nil
}

// Get is the metadata of a file.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (FileInfoResponse, error) {
	if d.Objects == nil {
		return FileInfoResponse{}, noStore()
	}
	info, err := d.Objects.Info(ctx, in.Name)
	if errors.Is(err, jetstream.ErrObjectNotFound) {
		return FileInfoResponse{}, endpoint.Errorf(ErrNotFound, "file %s not found", in.Name)
	}
	if err != nil {
		return FileInfoResponse{}, endpoint.Internal("failed to get file info", err)
	}

	return FileInfoResponse{
		Name:        info.Name,
		Sha256:      hexDigest(info.Digest),
		Size:        int(info.Size), //nolint:gosec // an object's size fits an int
		ContentType: contentType(info),
	}, nil
}

// Upload creates a file. A name that exists with the same content answers
// 200 and writes nothing; with other content it is ErrExists, and update
// replaces it.
func Upload(
	ctx context.Context,
	d Deps,
	in UploadInput,
) (FileUploadResponse, error) {
	if d.Objects == nil {
		return FileUploadResponse{}, noStore()
	}
	if protected(in.Name) {
		return FileUploadResponse{}, endpoint.Errorf(ErrProtected,
			"cannot write osapi file %s", in.Name)
	}
	w := write{name: in.Name, contentType: in.ContentType, data: in.File}

	existing, err := d.Objects.Info(ctx, in.Name)
	if err == nil && existing != nil {
		if existing.Digest == w.digest() {
			return w.result(false), endpoint.Answer(200)
		}

		return FileUploadResponse{}, endpoint.Errorf(ErrExists,
			"file %s exists with other content; replace it with file update", in.Name)
	}

	if err := w.store(ctx, d.Objects); err != nil {
		return FileUploadResponse{}, err
	}

	return w.result(true), nil
}

// Update replaces a file, or creates it when the name is new. The same
// content answers 200 and writes nothing.
func Update(
	ctx context.Context,
	d Deps,
	in UpdateInput,
) (FileUploadResponse, error) {
	if d.Objects == nil {
		return FileUploadResponse{}, noStore()
	}
	if protected(in.Name) {
		return FileUploadResponse{}, endpoint.Errorf(ErrProtected,
			"cannot write osapi file %s", in.Name)
	}
	w := write{name: in.Name, contentType: in.ContentType, data: in.File}

	existing, err := d.Objects.Info(ctx, in.Name)
	existed := err == nil && existing != nil
	if existed && existing.Digest == w.digest() {
		return w.result(false), endpoint.Answer(200)
	}

	if err := w.store(ctx, d.Objects); err != nil {
		return FileUploadResponse{}, err
	}
	if existed {
		return w.result(true), endpoint.Answer(200)
	}

	return w.result(true), nil
}

// Delete removes a file.
func Delete(
	ctx context.Context,
	d Deps,
	in DeleteInput,
) (struct{}, error) {
	if d.Objects == nil {
		return struct{}{}, noStore()
	}
	if protected(in.Name) {
		return struct{}{}, endpoint.Errorf(ErrProtected, "cannot delete osapi file %s", in.Name)
	}

	_, err := d.Objects.Info(ctx, in.Name)
	if errors.Is(err, jetstream.ErrObjectNotFound) {
		return struct{}{}, endpoint.Errorf(ErrNotFound, "file %s not found", in.Name)
	}
	if err != nil {
		return struct{}{}, endpoint.Internal("failed to get file info", err)
	}
	if err := d.Objects.Delete(ctx, in.Name); err != nil {
		return struct{}{}, endpoint.Internal("failed to delete file", err)
	}

	return struct{}{}, nil
}

// Stale is each deploy whose object changed after the deploy, or went: the
// state KV's SHA-256 of what an agent wrote against the object's now.
func Stale(
	ctx context.Context,
	d Deps,
	_ StaleInput,
) (StaleDeploymentsResponse, error) {
	if d.Objects == nil {
		return StaleDeploymentsResponse{}, noStore()
	}
	if d.States == nil {
		return StaleDeploymentsResponse{}, endpoint.Errorf(ErrNoState,
			"the file state KV is not configured: set nats.file_state.bucket on the controller")
	}
	keys, err := d.States.Keys(ctx)
	if err != nil {
		return StaleDeploymentsResponse{}, endpoint.Internal("failed to list file state keys", err)
	}

	stale := []StaleDeployment{}
	for _, key := range keys {
		entry, err := d.States.Get(ctx, key)
		if err != nil {
			continue
		}
		var state managedfile.State
		if json.Unmarshal(entry.Value(), &state) != nil || state.UndeployedAt != "" {
			continue
		}

		current := ""
		if content, err := d.Objects.GetBytes(ctx, state.ObjectName); err == nil {
			sum := sha256.Sum256(content)
			current = hex.EncodeToString(sum[:])
			if current == state.SHA256 {
				continue
			}
		}
		stale = append(stale, StaleDeployment{
			ObjectName:  state.ObjectName,
			Hostname:    hostnameOf(key),
			Provider:    providerOf(state.Path),
			Path:        state.Path,
			DeployedSha: state.SHA256,
			CurrentSha:  current,
			DeployedAt:  state.DeployedAt,
		})
	}

	return StaleDeploymentsResponse{Items: stale, TotalItems: len(stale)}, nil
}

// write is a file that upload or update writes.
type write struct {
	name, contentType string
	data              []byte
}

// digest is the Object Store digest of the file's content.
func (w write) digest() string {
	sum := sha256.Sum256(w.data)

	return "SHA-256=" + base64.URLEncoding.EncodeToString(sum[:])
}

// kind is how a deploy treats the file: raw unless the caller says.
func (w write) kind() string {
	if w.contentType == "" {
		return "raw"
	}

	return w.contentType
}

func (w write) result(
	changed bool,
) FileUploadResponse {
	sum := sha256.Sum256(w.data)

	return FileUploadResponse{
		Name:        w.name,
		Sha256:      hex.EncodeToString(sum[:]),
		Size:        len(w.data),
		Changed:     changed,
		ContentType: w.kind(),
	}
}

func (w write) store(
	ctx context.Context,
	objects Objects,
) error {
	meta := jetstream.ObjectMeta{
		Name:    w.name,
		Headers: nats.Header{contentTypeHeader: []string{w.kind()}},
	}
	if _, err := objects.Put(ctx, meta, bytes.NewReader(w.data)); err != nil {
		return endpoint.Internal("failed to store file", err)
	}

	return nil
}

func noStore() error {
	return endpoint.Errorf(ErrNoStore,
		"the Object Store is not configured: set nats.objects.bucket on the controller")
}

// protected is a name that osapi owns.
func protected(
	name string,
) bool {
	return strings.HasPrefix(name, "osapi/")
}

// hexDigest is an Object Store digest, SHA-256= and base64, as hex.
func hexDigest(
	digest string,
) string {
	b64 := strings.TrimPrefix(digest, "SHA-256=")
	raw, err := base64.URLEncoding.DecodeString(b64)
	if err != nil {
		return b64
	}

	return hex.EncodeToString(raw)
}

func contentType(
	info *jetstream.ObjectInfo,
) string {
	if info.Headers == nil {
		return ""
	}

	return info.Headers.Get(contentTypeHeader)
}

// hostnameOf is the hostname of a state key, <hostname>.<sha256 of path>:
// the key without its dot and 64 hex characters.
func hostnameOf(
	key string,
) string {
	if len(key) > 65 {
		return key[:len(key)-65]
	}

	return key
}

// providerOf is the provider that deploys to a path.
func providerOf(
	path string,
) string {
	switch {
	case strings.HasPrefix(path, "/etc/systemd/system/"):
		return "service"
	case strings.HasPrefix(path, "/usr/local/share/ca-certificates/"):
		return "certificate"
	case strings.HasPrefix(path, "/etc/cron.d/"),
		strings.HasPrefix(path, "/etc/cron.hourly/"),
		strings.HasPrefix(path, "/etc/cron.daily/"),
		strings.HasPrefix(path, "/etc/cron.weekly/"),
		strings.HasPrefix(path, "/etc/cron.monthly/"):
		return "cron"
	default:
		return "file"
	}
}
