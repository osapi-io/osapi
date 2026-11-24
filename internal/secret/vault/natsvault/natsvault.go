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

// Package natsvault is the nats type of vault, the one that osapi ships.
// NATS has no secret store, so the vault encrypts each value with
// AES-256-GCM under the vault key, with a new nonce each time, and keeps it
// in the JetStream KV bucket of nats.secrets. The KV keeps the earlier
// versions of a key, as the history of the bucket allows.
package natsvault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/secret/vault"
)

// TypeName is the name of the type, as the type of a vault names it.
const TypeName = "nats"

// Type is the nats type of vault, registered with package vault.
var Type = vault.Register(vault.Define(
	TypeName,
	"Each value encrypted with AES-256-GCM under the vault key, in the KV bucket of nats.secrets",
	check,
	build,
))

// check is an error when spec does not name its key in one place.
func check(
	spec vault.Spec,
) error {
	switch {
	case spec.KeyFile != "" && spec.KeyEnv != "":
		return errors.New("set key_file or key_env, not both")
	case spec.KeyFile == "" && spec.KeyEnv == "":
		return errors.New("a nats vault needs its key: set key_file or key_env")
	}

	return nil
}

// build opens a nats vault in the controller.
func build(
	ctx context.Context,
	env vault.Env,
	spec vault.Spec,
	_ Config,
) (vault.Vault, error) {
	return Open(ctx, env.NATS, env.Bucket, spec.Name, Keys{
		FS:        env.FS,
		File:      spec.KeyFile,
		Env:       spec.KeyEnv,
		LookupEnv: env.LookupEnv,
	})
}

// Open is the nats vault name, in bucket. It fails when the bucket is not
// there or the key does not load, so a controller with a vault it cannot
// use does not start.
func Open(
	ctx context.Context,
	kv KV,
	bucket string,
	name string,
	keys Keys,
) (*Vault, error) {
	if bucket == "" {
		return nil, errors.New("nats.secrets.bucket is empty: a nats vault needs its bucket")
	}
	if _, err := kv.KVStatus(ctx, bucket); err != nil {
		return nil, fmt.Errorf("open the bucket %s: %w: run osapi nats setup", bucket, err)
	}
	if _, err := keys.Load(); err != nil {
		return nil, err
	}

	return &Vault{name: name, kv: kv, bucket: bucket, keys: keys}, nil
}

// Name is the name of the vault.
func (v *Vault) Name() string {
	return v.name
}

// Get is the value of key, decrypted, and its metadata.
func (v *Vault) Get(
	ctx context.Context,
	key string,
) ([]byte, vault.Meta, error) {
	rec, _, err := v.read(ctx, key)
	if err != nil {
		return nil, vault.Meta{}, err
	}
	ring, err := v.keys.Load()
	if err != nil {
		return nil, vault.Meta{}, err
	}
	value, err := ring.open(rec.KeyID, rec.Nonce, rec.Data, v.aad(key))
	if err != nil {
		return nil, vault.Meta{}, fmt.Errorf("get %s/%s: %w", v.name, key, err)
	}

	return value, rec.meta(key), nil
}

// Put encrypts value under the current key and writes it as the next
// version of key. A Put at the same time as a different one fails, and
// does not overwrite it.
func (v *Vault) Put(
	ctx context.Context,
	key string,
	value []byte,
) (vault.Meta, error) {
	ring, err := v.keys.Load()
	if err != nil {
		return vault.Meta{}, err
	}
	rec, revision, err := v.read(ctx, key)
	if err != nil && !errors.Is(err, vault.ErrNotFound) {
		return vault.Meta{}, err
	}
	next := record{
		Version: rec.Version + 1,
		Changed: time.Now().UTC(),
		By:      vault.Actor(ctx),
	}
	next.KeyID, next.Nonce, next.Data = ring.seal(value, v.aad(key))
	if err := v.write(ctx, key, next, revision); err != nil {
		return vault.Meta{}, err
	}

	return next.meta(key), nil
}

// Delete deletes key. The bucket keeps its earlier versions, as its
// history allows.
func (v *Vault) Delete(
	ctx context.Context,
	key string,
) error {
	if _, _, err := v.read(ctx, key); err != nil {
		return err
	}
	if err := v.kv.KVDelete(ctx, v.bucket, v.subject(key)); err != nil {
		return fmt.Errorf("delete %s/%s: %w", v.name, key, err)
	}

	return nil
}

// List is the metadata of each key of the vault, sorted by key.
func (v *Vault) List(
	ctx context.Context,
) ([]vault.Meta, error) {
	names, err := v.kv.KVKeys(ctx, v.bucket, v.name+".>")
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", v.name, err)
	}
	out := make([]vault.Meta, 0, len(names))
	for _, name := range names {
		key := strings.TrimPrefix(name, v.name+".")
		rec, _, err := v.read(ctx, key)
		if errors.Is(err, vault.ErrNotFound) {
			// Deleted after the list.
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, rec.meta(key))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })

	return out, nil
}

// Rotate encrypts each value of the vault that is not under the current
// key of ring again, with that key, and is how many it changed. A value that
// changes at the same time is read again. The metadata of each value stays.
func (v *Vault) Rotate(
	ctx context.Context,
	ring Keyring,
) (int, error) {
	metas, err := v.List(ctx)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, m := range metas {
		did, err := v.rotateKey(ctx, ring, m.Key)
		if err != nil {
			return changed, err
		}
		if did {
			changed++
		}
	}

	return changed, nil
}

// rotateKey encrypts key again with the current key of ring. It tries
// again when a Put changes the key at the same time.
func (v *Vault) rotateKey(
	ctx context.Context,
	ring Keyring,
	key string,
) (bool, error) {
	for range 3 {
		rec, revision, err := v.read(ctx, key)
		if errors.Is(err, vault.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if rec.KeyID == ring.ID() {
			return false, nil
		}
		value, err := ring.open(rec.KeyID, rec.Nonce, rec.Data, v.aad(key))
		if err != nil {
			return false, fmt.Errorf("rotate %s/%s: %w", v.name, key, err)
		}
		rec.KeyID, rec.Nonce, rec.Data = ring.seal(value, v.aad(key))
		err = v.write(ctx, key, rec, revision)
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, jetstream.ErrKeyExists) {
			return false, err
		}
	}

	return false, fmt.Errorf(
		"rotate %s/%s: the key changed three times during the rotation",
		v.name,
		key,
	)
}

// read is the record of key and its revision in the bucket.
func (v *Vault) read(
	ctx context.Context,
	key string,
) (record, uint64, error) {
	if err := vault.CheckKey(key); err != nil {
		return record{}, 0, err
	}
	entry, err := v.kv.KVGet(ctx, v.bucket, v.subject(key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return record{}, 0, fmt.Errorf("%s/%s: %w", v.name, key, vault.ErrNotFound)
	}
	if err != nil {
		return record{}, 0, fmt.Errorf("get %s/%s: %w", v.name, key, err)
	}
	var rec record
	if err := json.Unmarshal(entry.Value(), &rec); err != nil {
		return record{}, 0, fmt.Errorf("get %s/%s: %w", v.name, key, err)
	}

	return rec, entry.Revision(), nil
}

// write writes rec to key: a new key when revision is 0, or else only
// when the key is still at revision.
func (v *Vault) write(
	ctx context.Context,
	key string,
	rec record,
	revision uint64,
) error {
	data, _ := json.Marshal(rec)
	var err error
	if revision == 0 {
		_, err = v.kv.KVCreate(ctx, v.bucket, v.subject(key), data)
	} else {
		_, err = v.kv.KVUpdate(ctx, v.bucket, v.subject(key), data, revision)
	}
	if err != nil {
		return fmt.Errorf("put %s/%s: %w", v.name, key, err)
	}

	return nil
}

// subject is the KV key of key: <vault>.<key>.
func (v *Vault) subject(
	key string,
) string {
	return v.name + "." + key
}

// aad is what binds a value to its vault and key.
func (v *Vault) aad(
	key string,
) string {
	return v.name + "/" + key
}

func (r record) meta(
	key string,
) vault.Meta {
	return vault.Meta{Key: key, Version: r.Version, Changed: r.Changed, By: r.By}
}
