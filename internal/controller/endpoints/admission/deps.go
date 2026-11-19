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

package admission

import (
	"context"
	"time"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/job"
)

// Deps is what the admission group's handlers use.
type Deps struct {
	// Store keeps the admission records. Nil when PKI is off; each method
	// then answers 503.
	Store Store
	// Jobs reads the registry, for which nodes are online, and the last
	// time each node reported.
	Jobs Jobs
}

// Store keeps the admission records.
type Store interface {
	List(ctx context.Context) ([]admission.Record, error)
	Get(ctx context.Context, machineID string) (admission.Record, error)
	Admit(ctx context.Context, machineID string, fingerprint string) error
	Reject(ctx context.Context, machineID string, reason string) error
	Revoke(ctx context.Context, machineID string) error
	Delete(ctx context.Context, machineID string) error
}

// Jobs is the part of the job client the admission group uses.
type Jobs interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
	AgentsLastSeen(ctx context.Context) (map[string]time.Time, error)
}
