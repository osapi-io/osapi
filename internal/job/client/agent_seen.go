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
	"fmt"
	"strings"
	"time"
)

// seenKVPrefix is the key prefix of each agent's last-seen time in the
// state bucket, which has no TTL, so the time outlives the registry entry.
const seenKVPrefix = "seen."

// WriteAgentSeen records now as when the agent with machineID last
// reported.
func (c *Client) WriteAgentSeen(
	ctx context.Context,
	machineID string,
) error {
	if c.stateBucket == "" {
		return fmt.Errorf("agent state bucket not configured")
	}
	data := []byte(time.Now().UTC().Format(time.RFC3339Nano))
	if _, err := c.natsClient.KVPut(ctx, c.stateBucket, seenKVPrefix+machineID, data); err != nil {
		return fmt.Errorf("write last seen: %w", err)
	}

	return nil
}

// AgentsLastSeen returns when each agent last reported, by machine ID. A
// value that does not read is left out.
func (c *Client) AgentsLastSeen(
	ctx context.Context,
) (map[string]time.Time, error) {
	seen := map[string]time.Time{}
	if c.stateBucket == "" {
		return seen, nil
	}
	keys, err := c.natsClient.KVKeys(ctx, c.stateBucket, seenKVPrefix+">")
	if err != nil {
		return nil, fmt.Errorf("list last seen: %w", err)
	}
	for _, key := range keys {
		entry, err := c.natsClient.KVGet(ctx, c.stateBucket, key)
		if err != nil {
			continue
		}
		var t time.Time
		if t.UnmarshalText(entry.Value()) != nil {
			continue
		}
		seen[strings.TrimPrefix(key, seenKVPrefix)] = t
	}

	return seen, nil
}
