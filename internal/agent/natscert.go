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

package agent

import (
	"bytes"
	"context"
	"encoding/pem"
	"log/slog"
	"path/filepath"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/atomicfile"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

// followNATSCert follows the rotation of the NATS server's certificate, from
// the record in the state bucket, at each heartbeat. The agent pins that
// certificate in its join's CA file, so the file holds what a record that
// a controller key signed names: during a rotation the current certificate
// and the next one, and after the switch the new one alone. The agent's
// NATS connection reads the file again at each handshake, so it trusts the
// new certificate when the server switches, with no restart. An agent on
// the NATS server's host that trusts the server's own nats.crt trusts what
// the switch renames into place, so it only reports the next certificate. A
// CA file of the agent's own config is the operator's, and the agent leaves
// it alone.
func (a *Agent) followNATSCert(
	ctx context.Context,
) {
	a.natsCertNext = ""
	file := filepath.Join(a.appConfig.Agent.PKI.KeyDir, natsJoinCA)
	serverCert, _ := config.SelfSignedFiles(a.appConfig.NATS.Server.KeyDir, "nats")
	ca := a.appConfig.Agent.NATS.TLS.CAFile
	if ca != file && ca != serverCert {
		return
	}
	rec, ok, err := a.jobClient.NATSCert(ctx)
	if err != nil {
		a.pkiLogger.Warn(
			"failed to read the NATS certificate record",
			slog.String("error", err.Error()),
		)

		return
	}
	if !ok {
		return
	}
	if !rec.SignedWith(a.pkiManager.VerifyWithGrace) {
		a.pkiLogger.Warn("the NATS certificate record is not signed by a controller key " +
			"that this agent has: the agent keeps the certificate that it trusts")

		return
	}
	if ca == serverCert {
		if rec.Rotating() {
			a.natsCertNext = pki.CertFingerprint(rec.Next)
		}

		return
	}
	want := append([]byte{}, rec.Current...)
	if rec.Rotating() {
		want = append(want, rec.Next...)
	}
	have, _ := avfs.ReadFile(a.appFs, file)
	if !samePEM(have, want) {
		if err := atomicfile.Write(ctx, a.appFs, file, want, 0o644); err != nil {
			a.pkiLogger.Warn(
				"failed to keep the NATS server's certificates",
				slog.String("error", err.Error()),
			)

			return
		}
		a.pkiLogger.Info(
			"the NATS server rotates its certificate: trusting its certificates",
			slog.String("current", pki.CertFingerprint(rec.Current)),
			slog.String("next", pki.CertFingerprint(rec.Next)),
		)
	}
	if rec.Rotating() {
		a.natsCertNext = pki.CertFingerprint(rec.Next)
	}
}

// samePEM reports whether a and b hold the same certificates, in order.
func samePEM(
	a []byte,
	b []byte,
) bool {
	var ab, bb []*pem.Block
	for block, rest := pem.Decode(a); block != nil; block, rest = pem.Decode(rest) {
		ab = append(ab, block)
	}
	for block, rest := pem.Decode(b); block != nil; block, rest = pem.Decode(rest) {
		bb = append(bb, block)
	}
	if len(ab) != len(bb) {
		return false
	}
	for i := range ab {
		if !bytes.Equal(ab[i].Bytes, bb[i].Bytes) {
			return false
		}
	}

	return true
}
