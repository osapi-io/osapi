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
	"regexp"
	"strconv"
)

// Update writes a file and keeps a record of it.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	f, err := file(in)
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
	return p.Read(ctx, in.Path)
}

// account is a user or group name chown can be given safely.
var account = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// file is in as a File: its mode parsed, and its owner and group checked,
// since they reach chown.
func file(
	in UpdateInput,
) (File, error) {
	f := File{
		Path:     in.Path,
		Content:  []byte(in.Content),
		Owner:    in.Owner,
		Group:    in.Group,
		Secret:   in.Secret,
		Metadata: in.Metadata,
	}
	if in.Mode != "" {
		mode, err := strconv.ParseUint(in.Mode, 8, 32)
		if err != nil || mode > 0o7777 {
			return File{}, fmt.Errorf("invalid mode %q: want octal, such as 0600", in.Mode)
		}
		f.Mode = fs.FileMode(mode)
	}
	for _, name := range []string{in.Owner, in.Group} {
		if name != "" && !account.MatchString(name) {
			return File{}, fmt.Errorf("invalid owner or group %q", name)
		}
	}

	return f, nil
}
