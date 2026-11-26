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

package natsauth

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsauth"
)

// Password makes a user and a password for each osapi role.
func Password(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	return create(ctx, d, in, "password", natsauth.Password)
}

// Operator makes an operator, an account and a credentials file for each
// osapi role.
func Operator(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	return create(ctx, d, in, "operator", natsauth.Operator)
}

// NKey makes an nkey for each osapi role.
func NKey(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	return create(ctx, d, in, "nkey", natsauth.NKey)
}

// create makes in.Out and writes the files of write to it.
func create(
	ctx context.Context,
	d Deps,
	in CreateInput,
	kind string,
	write func(context.Context, avfs.VFS, string, string, config.NATS) (natsauth.Result, error),
) (CreateOutput, error) {
	if err := d.FS.MkdirAll(in.Out, 0o700); err != nil {
		return CreateOutput{}, fmt.Errorf("make %s: %w", in.Out, err)
	}
	r, err := write(ctx, d.FS, in.Out, d.Config.Controller.NATS.Namespace, d.Config.NATS)
	if err != nil {
		return CreateOutput{}, err
	}
	out := CreateOutput{Files: []Written{}}
	for _, f := range r.Files {
		out.Files = append(out.Files, Written{Role: f.Role, Path: f.Path})
	}
	if in.Print {
		out.Server = r.Server
	}
	audit.RecordLocal(ctx, d.Audit, "natsauth.create", in.Out,
		fmt.Sprintf("type=%s files=%d", kind, len(out.Files)))

	return out, nil
}
