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

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// ControllerKey is the record of the rotation of the controller's key, in a
// pki.controller_key.get span. With no record, it is the zero record and
// false: the controller has not rotated its key, and has written none.
func (c *Client) ControllerKey(
	ctx context.Context,
) (pki.ControllerKeyRecord, bool, error) {
	ctx, span := opspan.Start(ctx, tracerName, "pki.controller_key.get")
	rec, ok, err := c.controllerKey(ctx)

	return rec, ok, opspan.End(span, err)
}

// controllerKey reads the record.
func (c *Client) controllerKey(
	ctx context.Context,
) (pki.ControllerKeyRecord, bool, error) {
	if c.stateBucket == "" {
		return pki.ControllerKeyRecord{}, false, fmt.Errorf("agent state bucket not configured")
	}
	entry, err := c.natsClient.KVGet(ctx, c.stateBucket, pki.ControllerKeyStateKey)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return pki.ControllerKeyRecord{}, false, nil
	}
	if err != nil {
		return pki.ControllerKeyRecord{}, false, fmt.Errorf(
			"read the controller key record: %w",
			err,
		)
	}
	var rec pki.ControllerKeyRecord
	if err := json.Unmarshal(entry.Value(), &rec); err != nil {
		return pki.ControllerKeyRecord{}, false, fmt.Errorf("the controller key record: %w", err)
	}

	return rec, true, nil
}
