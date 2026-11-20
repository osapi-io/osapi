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
	"context"

	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
	"github.com/osapi-io/osapi/internal/job"
)

// Deps is what the agent group's handlers use. The controller gives it at
// startup; a test gives fakes.
type Deps struct {
	// Jobs reads the registry and writes the drain flag and the timeline.
	Jobs Jobs
	// Health reports the health of each component, for status.
	Health health.Reporter
}

// Jobs is the part of the job client the agent group uses.
type Jobs interface {
	ListAgents(ctx context.Context) ([]job.AgentInfo, error)
	GetAgent(ctx context.Context, hostname string) (*job.AgentInfo, error)
	SetDrainFlag(ctx context.Context, hostname string) error
	DeleteDrainFlag(ctx context.Context, hostname string) error
	WriteAgentTimelineEvent(ctx context.Context, hostname, event, message string) error
}
