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

package agentkey

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/pki"
)

// Fingerprint loads the agent's keys, or makes each one that does not exist, and
// is the machine's ID, hostname, public encryption key and the fingerprint that
// covers both keys, and the command that admits it.
func Fingerprint(
	_ context.Context,
	d Deps,
	_ FingerprintInput,
) (FingerprintOutput, error) {
	mgr := pki.New(d.FS, d.KeyDir, "agent", pki.WithEncryptionKey())
	if err := mgr.LoadOrGenerate(); err != nil {
		return FingerprintOutput{}, fmt.Errorf("load the agent key from %s: %w", d.KeyDir, err)
	}
	ident, err := d.Identity()
	if err != nil {
		return FingerprintOutput{}, err
	}
	fp := mgr.Fingerprint()

	return FingerprintOutput{
		MachineID:     ident.MachineID,
		Hostname:      ident.Hostname,
		EncryptionKey: mgr.EncryptionPublicKey(),
		Fingerprint:   fp,
		Admit: fmt.Sprintf(
			"osapi client admission admit --machine-id %s --fingerprint %s", ident.MachineID, fp,
		),
	}, nil
}
