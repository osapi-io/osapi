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

package audit

import (
	"context"
	"net/http"
	"os/user"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// Actor is the kind of who made a change that did not go through the API.
type Actor string

// The actors of a change that did not go through the API. A request
// through the API has none: its User is the subject of its token.
const (
	// ActorSystem is osapi itself: auto_admit, delete_after.
	ActorSystem Actor = "system"
	// ActorJoinToken is a join token that admitted a machine.
	ActorJoinToken Actor = "join-token"
	// ActorLocal is a person who ran a command on the controller's host,
	// named by the user of the operating system.
	ActorLocal Actor = "local"
	// ActorAgent is a machine that asked for a change itself, named by its
	// machine ID: an enrollment request, a renewal of its NATS user, or a
	// hostname that it reported.
	ActorAgent Actor = "agent"
)

// Writer is where an entry goes: the audit stream. A *StreamStore is one.
type Writer interface {
	Write(ctx context.Context, entry Entry) error
}

// Change is the entry of a change that did not go through the API: actor
// and name (user) did action on target, with the outcome code, an HTTP
// status such as 200 or 403, and summary, which holds no secret.
func Change(
	ctx context.Context,
	actor Actor,
	name string,
	action string,
	target string,
	code int,
	summary string,
) Entry {
	e := Entry{
		ID:             uuid.New().String(),
		Timestamp:      time.Now(),
		User:           name,
		Roles:          []types.Role{},
		Actor:          actor,
		Action:         action,
		Target:         target,
		OperationID:    action,
		ResponseCode:   code,
		RequestSummary: summary,
	}
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		e.TraceID = sc.TraceID().String()
	}

	return e
}

// currentUser is the user of the operating system. Tests swap it.
var currentUser = user.Current

// LocalUser is the name of the user of the operating system who runs a
// local command, or "unknown".
func LocalUser() string {
	u, err := currentUser()
	if err != nil || u.Username == "" {
		return "unknown"
	}

	return u.Username
}

// Recorder writes an entry to the audit log of the controller. A local
// command gets one from its deps, or nil when the controller keeps no log.
type Recorder func(ctx context.Context, e Entry) error

// RecordLocal writes, with record, the entry of a change that a local
// command made, with the user of the operating system as the actor. A write
// that fails does not fail the command, which already made the change. A nil
// record writes nothing.
func RecordLocal(
	ctx context.Context,
	record Recorder,
	action string,
	target string,
	summary string,
) {
	if record == nil {
		return
	}
	_ = record(ctx, Change(ctx, ActorLocal, LocalUser(), action, target, http.StatusOK, summary))
}
