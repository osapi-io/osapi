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

package enrollment

import (
	"fmt"
	"log/slog"
)

// KeyRotationMessage is published to agents when the controller
// rotates its keypair.
type KeyRotationMessage struct {
	NewPublicKey []byte `json:"new_public_key"`
}

// RotateControllerKey publishes the controller's current public key
// to all agents via the pki.rotate subject. Agents receiving this
// message move their current controller key to previous and store
// the new key, enabling a grace period where both keys are accepted.
func (w *Watcher) RotateControllerKey() error {
	msg := KeyRotationMessage{
		NewPublicKey: w.pkiProvider.PublicKey(),
	}

	data, err := marshalFn(msg)
	if err != nil {
		return fmt.Errorf("marshal rotation message: %w", err)
	}

	subject := enrollSubject(w.namespace, "pki.rotate")
	if err := w.nc.PublishCore(subject, data); err != nil {
		return fmt.Errorf("publish key rotation: %w", err)
	}

	w.logger.Info(
		"published controller key rotation",
		slog.String("subject", subject),
	)

	return nil
}
