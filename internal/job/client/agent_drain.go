// Copyright (c) 2025 John Dewey

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
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

// CheckDrainFlag reports whether the drain flag exists for the given
// identifier. Callers pass a machine ID (hex string, no sanitization needed).
// Only a missing key is "no flag": a read that fails is an error, so a NATS
// outage does not read as an undrain.
func (c *Client) CheckDrainFlag(
	ctx context.Context,
	machineID string,
) (bool, error) {
	if c.stateBucket == "" {
		return false, nil
	}

	key := "drain." + machineID
	_, err := c.natsClient.KVGet(ctx, c.stateBucket, key)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read the drain flag of %s: %w", machineID, err)
	}

	return true, nil
}

// SetDrainFlag writes the drain flag for an agent in the state KV bucket.
// The agent detects this flag on heartbeat and stops accepting jobs.
// Callers pass a machine ID (hex string, no sanitization needed).
func (c *Client) SetDrainFlag(
	ctx context.Context,
	machineID string,
) error {
	if c.stateBucket == "" {
		return fmt.Errorf("agent state bucket not configured")
	}

	key := "drain." + machineID
	_, err := c.natsClient.KVPut(ctx, c.stateBucket, key, []byte("1"))
	if err != nil {
		return fmt.Errorf("set drain flag: %w", err)
	}

	c.logger.Debug(
		"set drain flag",
		slog.String("machine_id", machineID),
		slog.String("key", key),
	)

	return nil
}

// DeleteDrainFlag removes the drain flag for an agent from the state KV bucket.
// The agent detects this on heartbeat and resumes accepting jobs.
// Callers pass a machine ID (hex string, no sanitization needed).
func (c *Client) DeleteDrainFlag(
	ctx context.Context,
	machineID string,
) error {
	if c.stateBucket == "" {
		return fmt.Errorf("agent state bucket not configured")
	}

	key := "drain." + machineID
	err := c.natsClient.KVDelete(ctx, c.stateBucket, key)
	if err != nil {
		return fmt.Errorf("delete drain flag: %w", err)
	}

	c.logger.Debug(
		"deleted drain flag",
		slog.String("machine_id", machineID),
		slog.String("key", key),
	)

	return nil
}
