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

package agent_test

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/provider"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

type greeter interface{ Greet(name string) string }

type english struct{}

func (english) Greet(name string) string { return "hello " + name }

type greetIn struct {
	Name string `json:"name" validate:"omitempty,max=10"`
}

type greetOut struct {
	Text string `json:"text"`
}

func greet(
	_ context.Context,
	g greeter,
	in greetIn,
) (greetOut, error) {
	if in.Name == "boom" {
		return greetOut{}, errors.New("boom")
	}

	return greetOut{Text: g.Greet(in.Name)}, nil
}

func greetConcrete(
	_ context.Context,
	g english,
	in greetIn,
) (greetOut, error) {
	return greetOut{Text: g.Greet(in.Name)}, nil
}

type loginIn struct {
	Token secret.Value `json:"token"`
}

func login(
	_ context.Context,
	g greeter,
	in loginIn,
) (greetOut, error) {
	return greetOut{Text: g.Greet(in.Token.Reveal())}, nil
}

func build(
	p english,
	err error,
) func(host.Host) (english, error) {
	return func(host.Host) (english, error) { return p, err }
}

type ProcessorBehaviorPublicTestSuite struct {
	suite.Suite
}

func (s *ProcessorBehaviorPublicTestSuite) SetupTest() {
	behavior.Reset()
	behavior.Register(behavior.Define("greeting", "Says hello", behavior.Node,
		behavior.Platform("test", build(english{}, nil)),
		behavior.Read("get", greet),
		behavior.Read("list", greetConcrete),
		behavior.Read("show", greet).On("debian"),
		behavior.Read("login", login),
	))
	behavior.Register(behavior.Define("broken", "Fails to build", behavior.Node,
		behavior.Platform("test", build(english{}, errors.New("no socket"))),
		behavior.Read("get", greet),
	))
	behavior.Register(behavior.Define("elsewhere", "Another platform only", behavior.Node,
		behavior.Platform("darwin", build(english{}, nil)),
		behavior.Read("get", greet),
	))
	behavior.Register(behavior.Define("declined", "Declines this host", behavior.Node,
		behavior.Platform("test", build(english{}, provider.ErrUnsupported)),
		behavior.Read("get", greet),
	))
}

func (s *ProcessorBehaviorPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *ProcessorBehaviorPublicTestSuite) TestBuildBehaviorProviders() {
	providers := agent.BuildBehaviorProviders(host.Host{Platform: "test"}, slog.Default())

	s.Equal([]any{english{}}, agent.ExportBuiltProviders(providers))
}

func (s *ProcessorBehaviorPublicTestSuite) TestNewBehaviorProcessor() {
	tests := []struct {
		name            string
		operation       string
		data            string
		want            string
		wantErr         string
		wantUnsupported bool
	}{
		{
			name:      "when the provider implements the method's interface",
			operation: "greeting.get",
			data:      `{"name":"web"}`,
			want:      `{"text":"hello web"}`,
		},
		{
			name:      "when the method wants the concrete provider",
			operation: "greeting.list",
			data:      `{"name":"db"}`,
			want:      `{"text":"hello db"}`,
		},
		{
			name:      "when the operation has no method",
			operation: "greeting",
			wantErr:   "invalid behavior operation",
		},
		{
			name:      "when the behavior is unknown",
			operation: "nope.get",
			wantErr:   "unknown behavior",
		},
		{
			name:      "when the method is unknown",
			operation: "greeting.nope",
			wantErr:   "unknown method",
		},
		{
			name:      "when the provider failed to build",
			operation: "broken.get",
			wantErr:   "broken provider on test: no socket",
		},
		{
			name:            "when the behavior has no provider on this platform",
			operation:       "elsewhere.get",
			wantUnsupported: true,
		},
		{
			name:            "when the method runs only on another platform",
			operation:       "greeting.show",
			wantUnsupported: true,
		},
		{
			name:            "when the provider declines this host",
			operation:       "declined.get",
			wantUnsupported: true,
		},
		{
			name:      "when the job's input breaks a rule, the method does not run",
			operation: "greeting.get",
			data:      `{"name":"a name over ten"}`,
			wantErr:   "max",
		},
		{
			name:      "when the method fails",
			operation: "greeting.get",
			data:      `{"name":"boom"}`,
			wantErr:   "boom",
		},
	}

	process := agent.NewBehaviorProcessor(
		agent.BuildBehaviorProviders(host.Host{Platform: "test"}, slog.Default()),
		slog.Default(),
	)

	for _, tc := range tests {
		s.Run(tc.name, func() {
			out, err := process(context.Background(), job.Request{
				Operation: tc.operation,
				Data:      json.RawMessage(tc.data),
			})

			switch {
			case tc.wantUnsupported:
				s.ErrorIs(err, provider.ErrUnsupported)
			case tc.wantErr != "":
				s.ErrorContains(err, tc.wantErr)
			default:
				s.Require().NoError(err)
				s.JSONEq(tc.want, string(out))
			}
		})
	}
}

// TestSealedSecrets: the dispatch opens each sealed secret of the job with
// the Opener of the job, before the provider runs.
func (s *ProcessorBehaviorPublicTestSuite) TestSealedSecrets() {
	own, err := ecdh.P256().GenerateKey(rand.Reader)
	s.Require().NoError(err)
	pub := pki.EncodeEncryptionKey(own.PublicKey())
	sealer, err := seal.New("job-1")
	s.Require().NoError(err)
	key, err := sealer.KeyFor(pub)
	s.Require().NoError(err)
	opener, err := seal.Open("job-1", key, sealer.From(), own)
	s.Require().NoError(err)
	sealed := map[string]string{"m-1": key}

	tests := []struct {
		name    string
		ctx     context.Context
		data    string
		want    string
		wantErr error
		errText string
	}{
		{
			name: "when the job has a sealed secret, the provider gets its value",
			ctx:  seal.WithOpener(context.Background(), opener),
			data: `{"token":"` + sealer.Value("web") + `"}`,
			want: `{"text":"hello web"}`,
		},
		{
			name:    "when the job was not sealed to this agent, it does not run",
			ctx:     context.Background(),
			data:    `{"token":"` + sealer.Value("web") + `"}`,
			wantErr: seal.ErrNotSealedToAgent,
		},
		{
			name:    "when the input is not the method's, it is invalid",
			ctx:     seal.WithOpener(context.Background(), opener),
			data:    `{"token":1}`,
			wantErr: behavior.ErrInvalid,
		},
		{
			name:    "when a secret is not sealed, it does not run",
			ctx:     seal.WithOpener(context.Background(), opener),
			data:    `{"token":"web"}`,
			errText: "a secret of the job is not sealed",
		},
	}

	process := agent.NewBehaviorProcessor(
		agent.BuildBehaviorProviders(host.Host{Platform: "test"}, slog.Default()),
		slog.Default(),
	)
	for _, tc := range tests {
		s.Run(tc.name, func() {
			out, err := process(tc.ctx, job.Request{
				JobID:      "job-1",
				Operation:  "greeting.login",
				Data:       json.RawMessage(tc.data),
				SealedKeys: sealed,
				SealedFrom: sealer.From(),
			})

			switch {
			case tc.wantErr != nil:
				s.ErrorIs(err, tc.wantErr)
			case tc.errText != "":
				s.EqualError(err, tc.errText)
			default:
				s.Require().NoError(err)
				s.JSONEq(tc.want, string(out))
			}
		})
	}
}

func TestProcessorBehaviorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessorBehaviorPublicTestSuite))
}
