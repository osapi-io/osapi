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

package vaultkey

import (
	"context"
	"fmt"
	"os"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// Create writes a new key to the key_file of the nats vault in.Vault.
func Create(
	ctx context.Context,
	d Deps,
	in Input,
) (CreateOutput, error) {
	spec, err := natsSpec(d, in.Vault)
	if err != nil {
		return CreateOutput{}, err
	}
	ring, err := natsvault.CreateKey(ctx, d.FS, spec.KeyFile)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("vault %q: %w", spec.Name, err)
	}

	audit.RecordLocal(ctx, d.Audit, "vaultkey.create", spec.Name, "key_id="+ring.ID())

	return CreateOutput{Name: spec.Name, KeyFile: spec.KeyFile, KeyID: ring.ID()}, nil
}

// Rotate adds a new key to the key_file of the nats vault in.Vault, encrypts
// each value of the vault again with it, and drops the old key.
func Rotate(
	ctx context.Context,
	d Deps,
	in Input,
) (RotateOutput, error) {
	spec, err := natsSpec(d, in.Vault)
	if err != nil {
		return RotateOutput{}, err
	}
	kv, closeKV, err := d.Dial(ctx)
	if err != nil {
		return RotateOutput{}, err
	}
	defer closeKV()

	bucket := ""
	if d.Config.NATS.Secrets.Bucket != "" {
		bucket = job.ApplyNamespaceToInfraName(
			d.Config.Controller.NATS.Namespace, d.Config.NATS.Secrets.Bucket,
		)
	}
	v, err := natsvault.Open(ctx, kv, bucket, spec.Name, natsvault.Keys{
		FS:        d.FS,
		File:      spec.KeyFile,
		Env:       spec.KeyEnv,
		LookupEnv: os.LookupEnv,
	})
	if err != nil {
		return RotateOutput{}, fmt.Errorf("vault %q: %w", spec.Name, err)
	}
	ring, n, err := natsvault.RotateKey(ctx, v)
	if err != nil {
		return RotateOutput{}, fmt.Errorf("vault %q: %w", spec.Name, err)
	}

	audit.RecordLocal(ctx, d.Audit, "vaultkey.rotate", spec.Name,
		fmt.Sprintf("key_id=%s values=%d", ring.ID(), n))

	return RotateOutput{Name: spec.Name, KeyFile: spec.KeyFile, KeyID: ring.ID(), Values: n}, nil
}

// natsSpec is the vault name of the config, which must be of type nats.
func natsSpec(
	d Deps,
	name string,
) (vault.Spec, error) {
	for _, spec := range d.Config.Controller.Secrets.Vaults {
		if spec.Name != name {
			continue
		}
		if spec.Type != natsvault.TypeName {
			return vault.Spec{}, fmt.Errorf(
				"vault %q is of type %s: only a nats vault has a key of osapi",
				name, spec.Type,
			)
		}

		return spec, nil
	}

	return vault.Spec{}, fmt.Errorf(
		"no vault %q in controller.secrets.vaults of the config", name,
	)
}
