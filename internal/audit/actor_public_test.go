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

package audit_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os/user"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/audit"
)

type ActorPublicTestSuite struct {
	suite.Suite
}

func TestActorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ActorPublicTestSuite))
}

func (s *ActorPublicTestSuite) TestChange() {
	e := audit.Change(
		context.Background(),
		audit.ActorSystem,
		"auto_admit",
		"agent.admit",
		"m-1",
		200,
		"hostname=web-01",
	)
	s.NotEmpty(e.ID)
	s.False(e.Timestamp.IsZero())
	s.Equal(audit.ActorSystem, e.Actor)
	s.Equal("auto_admit", e.User)
	s.Equal("agent.admit", e.Action)
	s.Equal("agent.admit", e.OperationID, "the list filters by operation")
	s.Equal("m-1", e.Target)
	s.Equal(200, e.ResponseCode)
	s.Empty(e.TraceID)
	s.NotNil(e.Roles, "an empty list: the response's roles are never null")

	ctx := trace.ContextWithSpanContext(
		context.Background(),
		trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1},
		}),
	)
	s.Equal(
		trace.TraceID{1}.String(),
		audit.Change(ctx, audit.ActorLocal, "ops", "x", "y", 200, "").TraceID,
	)
}

func (s *ActorPublicTestSuite) TestLocalUser() {
	defer audit.SetCurrentUser(
		func() (*user.User, error) { return &user.User{Username: "ops"}, nil },
	)()
	s.Equal("ops", audit.LocalUser())

	defer audit.SetCurrentUser(func() (*user.User, error) { return nil, errors.New("no passwd") })()
	s.Equal("unknown", audit.LocalUser())
}

func (s *ActorPublicTestSuite) TestRecordLocal() {
	defer audit.SetCurrentUser(
		func() (*user.User, error) { return &user.User{Username: "ops"}, nil },
	)()
	var logs bytes.Buffer
	defer slog.SetDefault(slog.Default())
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	var got []audit.Entry
	audit.RecordLocal(context.Background(), func(_ context.Context, e audit.Entry) error {
		got = append(got, e)

		return errors.New("the stream is down")
	}, "token.generate", "admin@example.com", "roles=[admin]")
	s.Require().Len(got, 1, "a failed write does not panic or retry")
	s.Equal(audit.ActorLocal, got[0].Actor)
	s.Equal("ops", got[0].User)
	s.Equal("token.generate", got[0].Action)
	s.Equal("admin@example.com", got[0].Target)
	s.Contains(logs.String(), "the audit entry of a local command was not written")
	s.Contains(logs.String(), "the stream is down")

	s.NotPanics(func() { audit.RecordLocal(context.Background(), nil, "x", "y", "") })
}
