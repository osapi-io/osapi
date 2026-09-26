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
	"context"
	"errors"

	"github.com/osapi-io/osapi/internal/job/client"
)

// clientKeyStore adapts the watcher's store to the narrow lookup the job
// client depends on. The job client stays free of this package: the
// dependency runs from the controller to the shared client, not back.
type clientKeyStore struct {
	watcher *Watcher
}

// KeyStore returns the watcher as the job client's agent key store, so
// response verification reads the same records acceptance writes and shares
// their cache.
func (w *Watcher) KeyStore() client.AgentKeyStore {
	return clientKeyStore{watcher: w}
}

// LookupAgentKey implements client.AgentKeyStore, translating this package's
// causes into the client's. The distinction between them is preserved: a
// missing record and an unreadable store never collapse into one answer.
func (s clientKeyStore) LookupAgentKey(
	ctx context.Context,
	machineID string,
) (*client.AgentKey, error) {
	record, err := s.watcher.LookupAgentKey(ctx, machineID)

	switch {
	case errors.Is(err, ErrAgentKeyNotFound):
		return nil, client.ErrResponseKeyUnknown
	case err != nil:
		// Every other failure is a store that could not be read. It is
		// never reported as "no stored key", which would read as an agent
		// that has simply not enrolled yet.
		return nil, client.ErrResponseStoreUnavailable
	}

	return &client.AgentKey{
		MachineID: record.MachineID,
		Hostname:  record.Hostname,
		PublicKey: record.PublicKey,
	}, nil
}
