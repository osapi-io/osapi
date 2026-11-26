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

package secret

import (
	"context"
	"errors"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/secret/vault"
)

// VaultList is each vault of the controller, with its type, its targets
// and whether it answers a list of its keys.
func VaultList(
	ctx context.Context,
	d Deps,
	_ VaultListInput,
) (VaultListResponse, error) {
	items := []VaultInfo{}
	if d.Vaults != nil {
		for _, name := range d.Vaults.Names() {
			spec, _ := d.Vaults.Spec(name)
			v, _ := d.Vaults.Get(name)
			_, err := v.List(ctx)
			targets := spec.Targets
			if targets == nil {
				targets = []string{}
			}
			items = append(items, VaultInfo{
				Name:    name,
				Type:    spec.Type,
				Targets: targets,
				Answers: err == nil,
			})
		}
	}

	return VaultListResponse{Items: items, TotalItems: len(items)}, nil
}

// List is the metadata of each key of a vault, and no value.
func List(
	ctx context.Context,
	d Deps,
	in ListInput,
) (ListResponse, error) {
	v, err := open(d, in.Vault)
	if err != nil {
		return ListResponse{}, err
	}
	metas, err := v.List(ctx)
	if err != nil {
		return ListResponse{}, endpoint.Internal("failed to list the keys of the vault", err)
	}
	items := make([]Info, 0, len(metas))
	for _, m := range metas {
		items = append(items, info(in.Vault, m))
	}

	return ListResponse{Items: items, TotalItems: len(items)}, nil
}

// Get is the metadata of one key, and not its value.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (Info, error) {
	v, err := open(d, in.Vault)
	if err != nil {
		return Info{}, err
	}
	m, err := find(ctx, v, in.Vault, in.Key)
	if err != nil {
		return Info{}, err
	}

	return info(in.Vault, m), nil
}

// Reveal is the value of one key: the one method that gives a value. The
// route checks secret:reveal; the audit log records the call by its path,
// and the value goes only into the answer.
func Reveal(
	ctx context.Context,
	d Deps,
	in RevealInput,
) (Value, error) {
	v, err := open(d, in.Vault)
	if err != nil {
		return Value{}, err
	}
	value, _, err := v.Get(ctx, in.Key)
	if errors.Is(err, vault.ErrNotFound) {
		return Value{}, endpoint.Errorf(ErrNoKey, "vault %s has no key %s", in.Vault, in.Key)
	}
	if err != nil {
		return Value{}, endpoint.Internal("failed to read the key", err)
	}

	return Value{Value: string(value)}, nil
}

// Update writes the value as the next version of the key, and records the
// subject of the token as who changed it. The answer is the metadata.
func Update(
	ctx context.Context,
	d Deps,
	in UpdateInput,
) (Info, error) {
	v, err := open(d, in.Vault)
	if err != nil {
		return Info{}, err
	}
	ctx = vault.WithActor(ctx, authtoken.SubjectFrom(ctx))
	m, err := v.Put(ctx, in.Key, []byte(in.Body.Value.Reveal()))
	if err != nil {
		return Info{}, endpoint.Internal("failed to write the key", err)
	}

	return info(in.Vault, m), nil
}

// Delete deletes the key from the vault.
func Delete(
	ctx context.Context,
	d Deps,
	in DeleteInput,
) (struct{}, error) {
	v, err := open(d, in.Vault)
	if err != nil {
		return struct{}{}, err
	}
	err = v.Delete(ctx, in.Key)
	if errors.Is(err, vault.ErrNotFound) {
		return struct{}{}, endpoint.Errorf(ErrNoKey, "vault %s has no key %s", in.Vault, in.Key)
	}
	if err != nil {
		return struct{}{}, endpoint.Internal("failed to delete the key", err)
	}

	return struct{}{}, nil
}

// open is the vault with the name, or ErrNoVault.
func open(
	d Deps,
	name string,
) (vault.Vault, error) {
	if d.Vaults != nil {
		if v, ok := d.Vaults.Get(name); ok {
			return v, nil
		}
	}

	return nil, endpoint.Errorf(ErrNoVault,
		"no vault named %s: controller.secrets.vaults names each vault", name)
}

// find is the metadata of key from the list of the vault, so that Get
// never reads a value. A key that is not there is ErrNoKey.
func find(
	ctx context.Context,
	v vault.Vault,
	vaultName string,
	key string,
) (vault.Meta, error) {
	metas, err := v.List(ctx)
	if err != nil {
		return vault.Meta{}, endpoint.Internal("failed to list the keys of the vault", err)
	}
	for _, m := range metas {
		if m.Key == key {
			return m, nil
		}
	}

	return vault.Meta{}, endpoint.Errorf(ErrNoKey, "vault %s has no key %s", vaultName, key)
}

// info is the metadata of a key, as the API gives it.
func info(
	vaultName string,
	m vault.Meta,
) Info {
	return Info{
		Vault:   vaultName,
		Key:     m.Key,
		Version: int64(m.Version), //nolint:gosec // a version is a count of updates
		Changed: m.Changed,
		By:      m.By,
	}
}
