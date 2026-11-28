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

package managedfile

import (
	"context"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strconv"

	"github.com/osapi-io/osapi/internal/validation"
)

// Unapplied is the key of the metadata that marks a file whose write landed
// and whose next step did not finish, such as netplan apply or a reload of
// systemd. No caller writes it, so the record then differs from each write
// of the file: the next write of the same content changes the record and
// reports changed, and its caller runs the step again.
const Unapplied = "osapi.unapplied"

// MarkUnapplied is in with Unapplied in its metadata. A caller whose step
// after a write failed writes it, so a retry runs the step again.
func MarkUnapplied(
	in UpdateInput,
) UpdateInput {
	metadata := maps.Clone(in.Metadata)
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata[Unapplied] = "true"
	in.Metadata = metadata

	return in
}

// Update writes a file and keeps a record of it.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	if err := absolute(in.Path); err != nil {
		return Result{}, err
	}
	f, err := file(ctx, in)
	if err != nil {
		return Result{}, err
	}
	changed, err := p.Write(ctx, f)
	if err != nil {
		return Result{}, err
	}

	return Result{Path: in.Path, Changed: changed}, nil
}

// Delete removes a file osapi manages and records that it no longer does.
func Delete(
	ctx context.Context,
	p Provider,
	in DeleteInput,
) (Result, error) {
	if err := absolute(in.Path); err != nil {
		return Result{}, err
	}
	changed, err := p.Remove(ctx, in.Path)
	if err != nil {
		return Result{}, err
	}

	return Result{Path: in.Path, Changed: changed}, nil
}

// Get reads the record osapi keeps of a file.
func Get(
	ctx context.Context,
	p Provider,
	in GetInput,
) (Record, error) {
	if err := absolute(in.Path); err != nil {
		return Record{}, err
	}

	return p.Read(ctx, in.Path)
}

// absolute refuses a relative path, which would resolve against the agent's
// working directory, as the API does before a job is queued.
func absolute(
	path string,
) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("invalid path %q: want an absolute path", path)
	}

	return nil
}

// file is in as a File: its mode parsed, and its owner and group checked,
// since they reach chown.
func file(
	ctx context.Context,
	in UpdateInput,
) (File, error) {
	f := File{
		Path:         in.Path,
		Content:      []byte(in.Content),
		Owner:        in.Owner,
		Group:        in.Group,
		Secret:       in.Secret,
		Metadata:     in.Metadata,
		Object:       in.Object,
		ContentType:  in.ContentType,
		ObjectSHA256: in.ObjectSHA256,
	}
	if in.Mode != "" {
		mode, err := strconv.ParseUint(in.Mode, 8, 32)
		if err != nil || mode > 0o7777 {
			return File{}, fmt.Errorf("invalid mode %q: want octal, such as 0600", in.Mode)
		}
		f.Mode = fs.FileMode(mode)
	}
	// The owner and the group reach chown, so they follow the account_name
	// rule of the input, the one rule of an account name.
	for _, name := range []string{in.Owner, in.Group} {
		if _, ok := validation.Var(ctx, name, "omitempty,account_name"); !ok {
			return File{}, fmt.Errorf("invalid owner or group %q", name)
		}
	}

	return f, nil
}
