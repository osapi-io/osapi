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
	"errors"
	"fmt"
)

// Load is the keyring of the file, or of the environment variable when
// there is no file.
func (k Keys) Load() (Keyring, error) {
	if k.File != "" {
		data, err := k.FS.ReadFile(k.File)
		if err != nil {
			return Keyring{}, fmt.Errorf(
				"read the vault key: %w: make it with osapi controller vault key create",
				err,
			)
		}
		ring, err := ParseKeyring(data)
		if err != nil {
			return Keyring{}, fmt.Errorf("the vault key %s: %w", k.File, err)
		}

		return ring, nil
	}
	if k.Env == "" {
		return Keyring{}, errors.New("the vault has no key: set key_file or key_env")
	}
	data, ok := k.LookupEnv(k.Env)
	if !ok {
		return Keyring{}, fmt.Errorf("the vault key: %s is not set", k.Env)
	}
	ring, err := ParseKeyring([]byte(data))
	if err != nil {
		return Keyring{}, fmt.Errorf("the vault key in %s: %w", k.Env, err)
	}

	return ring, nil
}
