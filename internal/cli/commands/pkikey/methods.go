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

package pkikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"


	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/secret/vault/natsvault"
)

// jetStreamKeyBytes is the length of a JetStream key before it is hex
// encoded.
const jetStreamKeyBytes = 32

// randRead fills a key with random bytes (injectable for testing).
var randRead = rand.Read

// Errors of the key commands.
var (
	ErrNoKey    = errors.New("no key with this name: list shows each")
	ErrNotMade  = errors.New("not made")
	ErrExists   = errors.New("the key exists")
	ErrNoRotate = errors.New("osapi does not rotate this key")
)

// List is each key of osapi on this host, with its state.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	return ListOutput{Items: Keys(d)}, nil
}

// Keys is each key of osapi on this host. A key that cannot be read is in
// the list too, with its state.
func Keys(
	d Deps,
) []Key {
	srcs := sources(d)
	out := make([]Key, 0, len(srcs))
	for _, s := range srcs {
		out = append(out, read(d, s))
	}

	return out
}

// Get is one key in full. The agent's signing key has the command that
// admits this machine.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (Key, error) {
	s, err := find(d, in.Name)
	if err != nil {
		return Key{}, err
	}
	k := read(d, s)
	if k.State == StateNotMade {
		return Key{}, fmt.Errorf("%s is not on this host: %s; make it with "+
			"osapi pki key create --name %s: %w", s.name, s.component, s.name, ErrNotMade)
	}
	if s.name == agentName && k.State == StateOK && k.Fingerprint != "" && d.Identity != nil {
		if ident, err := d.Identity(); err == nil {
			k.Admit = fmt.Sprintf("osapi client admission admit --machine-id %s --fingerprint %s",
				ident.MachineID, k.Fingerprint)
		}
	}

	return k, nil
}

// Create makes the key with the name. It refuses a key that is there: a new
// key in its place is a rotation, and only a vault key rotates.
func Create(
	ctx context.Context,
	d Deps,
	in CreateInput,
) (Key, error) {
	s, err := find(d, in.Name)
	if err != nil {
		return Key{}, err
	}
	if s.typ == typeAgentEKey {
		// The two keys of the agent are made together.
		s, _ = find(d, agentName)
	}
	switch k := read(d, s); {
	case k.State == StateNotMade, k.State == StateOff:
	case k.State == StateOK && s.name == agentName && halfAgent(d):
		// An agent made before it had an encryption key gets it now.
	case k.State == StateOK:
		return Key{}, fmt.Errorf("%s: %w: %s", s.name, ErrExists, rotateHint(s))
	default:
		return Key{}, fmt.Errorf("%s is %s: osapi does not replace a key that it cannot read",
			s.name, k.State)
	}
	switch s.typ {
	case typeSigning:
		opts := []pki.Option{}
		if s.name == agentName {
			opts = append(opts, pki.WithEncryptionKey())
		}
		if err := pki.New(d.FS, s.dir, s.prefix, opts...).LoadOrGenerate(); err != nil {
			return Key{}, fmt.Errorf("make the %s key in %s: %w", s.name, s.dir, err)
		}
	case typeJetStream:
		if err := createJetStream(ctx, d, s.file); err != nil {
			return Key{}, err
		}
	default:
		if s.file == "" {
			return Key{}, fmt.Errorf("vault %q has no key_file: set it to the file to write, "+
				"or put the key in its key_env", strings.TrimPrefix(s.name, vaultPrefix))
		}
		if _, err := natsvault.CreateKey(ctx, d.FS, s.file); err != nil {
			return Key{}, fmt.Errorf("%s: %w", s.name, err)
		}
	}
	k := read(d, s)
	audit.RecordLocal(ctx, d.Audit, "key.create", s.name, summary(k))

	return k, nil
}

// Rotate makes a new key of a vault, encrypts each value of the vault again
// with it, and drops the old key. Any other key is refused, with what to do
// instead.
func Rotate(
	ctx context.Context,
	d Deps,
	in RotateInput,
) (RotateOutput, error) {
	s, err := find(d, in.Name)
	if err != nil {
		return RotateOutput{}, err
	}
	if s.typ != typeVault {
		return RotateOutput{}, fmt.Errorf("%s: %w: %s", s.name, ErrNoRotate, rotateHint(s))
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
	name := strings.TrimPrefix(s.name, vaultPrefix)
	v, err := natsvault.Open(ctx, kv, bucket, name, natsvault.Keys{
		FS: d.FS, File: s.file, Env: s.env, LookupEnv: d.LookupEnv,
	})
	if err != nil {
		return RotateOutput{}, fmt.Errorf("vault %q: %w", name, err)
	}
	_, n, err := natsvault.RotateKey(ctx, v)
	if err != nil {
		return RotateOutput{}, fmt.Errorf("vault %q: %w", name, err)
	}
	k := read(d, s)
	audit.RecordLocal(
		ctx,
		d.Audit,
		"key.rotate",
		s.name,
		fmt.Sprintf("%s values=%d", summary(k), n),
	)

	return RotateOutput{Key: k, Values: n}, nil
}

// find is the key named name.
func find(
	d Deps,
	name string,
) (source, error) {
	for _, s := range sources(d) {
		if s.name == name {
			return s, nil
		}
	}

	return source{}, fmt.Errorf("%s: %w", name, ErrNoKey)
}

// rotateHint is what to do instead of a new key in place of s.
func rotateHint(
	s source,
) string {
	switch s.typ {
	case typeVault:
		return "osapi pki key rotate --name " + s.name + " makes a new one and " +
			"encrypts each value again with it"
	case typeJetStream:
		return "a new key makes each stream and bucket of nats.server.store_dir " +
			"unreadable; keep this key, and a copy of it apart from the store"
	default:
		if s.name == controllerName {
			return "each admitted agent verifies the jobs of the controller with the " +
				"key it got at its admission; a new key needs each agent admitted again"
		}

		return "the controller admitted the fingerprint of these keys; a new key is " +
			"refused until the machine is admitted again: remove its admission, then " +
			"let the agent join again"
	}
}

// summary is what the audit log records of k: its ID or fingerprint, never
// the key.
func summary(
	k Key,
) string {
	if k.KeyID != "" {
		return "key_id=" + k.KeyID
	}

	return "fingerprint=" + k.Fingerprint
}

// halfAgent reports whether the agent has its signing key and not its
// encryption key, as an agent made before the encryption key had. Create
// makes the encryption key then.
func halfAgent(
	d Deps,
) bool {
	x, _ := find(d, "agent-encryption")

	return read(d, x).State == StateNotMade
}

// read is the key of s as list shows it.
func read(
	d Deps,
	s source,
) Key {
	k := Key{
		Name: s.name, Kind: s.kind, Component: s.component, File: s.display(),
		Algorithm: s.algorithm, State: StateOK,
	}
	var err error
	switch s.typ {
	case typeSigning:
		err = readSigning(d, s, &k)
	case typeAgentEKey:
		k.PublicKey, err = readEKey(d, s.files()[0])
	case typeJetStream:
		if s.file == "" {
			k.State, k.File = StateOff, ""

			return k
		}
		err = readJetStream(d, s.file, &k)
	default:
		err = readVault(d, s, &k)
	}
	if err != nil {
		k.State = state(err)
	}

	return k
}

// readSigning reads the Ed25519 key pair of s: its fingerprint, and for the
// agent the fingerprint that covers its encryption key too, as its
// admission does.
func readSigning(
	d Deps,
	s source,
	k *Key,
) error {
	files := s.files()
	_, privErr := d.FS.Stat(files[0])
	pubData, pubErr := d.FS.ReadFile(files[1])
	if errors.Is(privErr, fs.ErrNotExist) && errors.Is(pubErr, fs.ErrNotExist) {
		return pubErr
	}
	for _, f := range []struct {
		path string
		err  error
	}{{files[0], privErr}, {files[1], pubErr}} {
		if errors.Is(f.err, fs.ErrNotExist) {
			return fmt.Errorf("half a key pair: %s is not there", f.path)
		}
		if f.err != nil {
			return f.err
		}
	}
	pub, err := pki.ParsePublicKeyPEM(pubData)
	if err != nil {
		return fmt.Errorf("%s: %w", files[1], err)
	}
	k.Fingerprint = pki.FingerprintOf(pub)
	if s.name == agentName {
		x, _ := find(d, "agent-encryption")
		if xpub, err := readEKey(d, x.files()[0]); err == nil {
			k.Fingerprint = pki.FingerprintOfKeys(pub, xpub)
		}
	}

	return nil
}

// readEKey is the public P-256 key of the encryption key in file, in
// pki.EncryptionKeyEncoding.
func readEKey(
	d Deps,
	file string,
) (string, error) {
	data, err := d.FS.ReadFile(file)
	if err != nil {
		return "", err
	}
	key, err := pki.ParseEncryptionKeyFile(data)
	if err != nil {
		return "", fmt.Errorf("%s: %w", file, err)
	}

	return pki.EncodeEncryptionKey(key.PublicKey()), nil
}

// readJetStream reads the JetStream key in file, and is its ID.
func readJetStream(
	d Deps,
	file string,
	k *Key,
) error {
	data, err := d.FS.ReadFile(file)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return fmt.Errorf("%s is empty", file)
	}
	k.KeyID = keyID([]byte(key))

	return nil
}

// readVault reads the keyring of a vault, and is the ID of its current key
// and the number of its keys.
func readVault(
	d Deps,
	s source,
	k *Key,
) error {
	var data []byte
	switch {
	case s.file != "":
		b, err := d.FS.ReadFile(s.file)
		if err != nil {
			return err
		}
		data = b
	case s.env != "":
		v, ok := d.LookupEnv(s.env)
		if !ok {
			return fmt.Errorf("%w: %s is not set", fs.ErrNotExist, s.env)
		}
		data = []byte(v)
	default:
		return fmt.Errorf("%w: the vault sets no key_file and no key_env", fs.ErrNotExist)
	}
	ring, err := natsvault.ParseKeyring(data)
	if err != nil {
		return err
	}
	k.KeyID, k.Keys = ring.ID(), ring.Len()

	return nil
}

// keyID is the ID of a secret key: the first 16 hex digits of its SHA-256,
// as a vault records the ID of its key.
func keyID(
	key []byte,
) string {
	sum := sha256.Sum256(key)

	return hex.EncodeToString(sum[:8])
}

// state is why a key could not be read: not made when its file is not
// there, or unreadable with the reason.
func state(
	err error,
) string {
	if errors.Is(err, fs.ErrNotExist) {
		return StateNotMade
	}

	return "unreadable: " + err.Error()
}

// createJetStream writes a new random key to file, with mode 0600, and its
// directory with mode 0700. Create refuses a file that is there before it
// calls this, because a different key makes the store unreadable.
func createJetStream(
	ctx context.Context,
	d Deps,
	file string,
) error {
	if file == "" {
		return errors.New("nats.server.jetstream.key_file is not set: set it to the file to write")
	}
	raw := make([]byte, jetStreamKeyBytes)
	if _, err := randRead(raw); err != nil {
		return fmt.Errorf("make the key: %w", err)
	}
	// A directory that cannot be made fails the write, which says so.
	_ = d.FS.MkdirAll(filepath.Dir(file), 0o700)
	if err := atomicfile.Write(
		ctx,
		d.FS,
		file,
		[]byte(hex.EncodeToString(raw)+"\n"),
		0o600,
	); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}

	return nil
}
