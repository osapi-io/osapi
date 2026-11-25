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

package natsvault

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/atomicfile"
)

// keyMode is the mode of a key file: only its owner reads it.
const keyMode fs.FileMode = 0o600

// CreateKey writes a new vault key to file, and is its keyring. It does
// not replace a key that is there, because each value under it would then
// not decrypt: rotate a key with RotateKey.
func CreateKey(
	ctx context.Context,
	vfs avfs.VFS,
	file string,
) (Keyring, error) {
	if file == "" {
		return Keyring{}, errors.New("the vault has no key_file: set it to the file to write")
	}
	if _, err := vfs.Stat(file); err == nil {
		return Keyring{}, fmt.Errorf(
			"%s exists: rotate the key with osapi controller secret key rotate", file,
		)
	}
	if err := vfs.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return Keyring{}, fmt.Errorf("make the directory of %s: %w", file, err)
	}
	ring := NewKeyring(Keyring{})
	if err := atomicfile.Write(ctx, vfs, file, ring.Marshal(), keyMode); err != nil {
		return Keyring{}, fmt.Errorf("write %s: %w", file, err)
	}

	return ring, nil
}

// RotateKey adds a new key to the key file of v, encrypts each value of v
// again with it, then drops the old keys from the file. It is the new
// keyring, and how many values it encrypted again. A vault reads the file
// at each call, so a controller that runs puts each new value under the new
// key once the file has it. A rotation that stops leaves the old keys in
// the file, and a second rotation completes it.
func RotateKey(
	ctx context.Context,
	v *Vault,
) (Keyring, int, error) {
	if v.keys.File == "" {
		return Keyring{}, 0, fmt.Errorf(
			"the key of %s is in %s, not a file: rotate writes key_file",
			v.name, v.keys.Env,
		)
	}
	old, err := v.keys.Load()
	if err != nil {
		return Keyring{}, 0, err
	}
	ring := NewKeyring(old)
	if err := atomicfile.Write(ctx, v.keys.FS, v.keys.File, ring.Marshal(), keyMode); err != nil {
		return Keyring{}, 0, fmt.Errorf("write %s: %w", v.keys.File, err)
	}
	changed, err := v.Rotate(ctx, ring)
	if err != nil {
		return Keyring{}, changed, fmt.Errorf(
			"%w: %s keeps the old keys, so each value decrypts: run the rotation again",
			err, v.keys.File,
		)
	}
	current := ring.Current()
	if err := atomicfile.Write(
		ctx,
		v.keys.FS,
		v.keys.File,
		current.Marshal(),
		keyMode,
	); err != nil {
		return Keyring{}, changed, fmt.Errorf("write %s: %w", v.keys.File, err)
	}

	return current, changed, nil
}
