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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
)

// BehaviorProviders are the providers the agent built for every registered
// behavior on its platform, as each declares them with behavior.Platform.
type BehaviorProviders struct {
	platform behavior.PlatformName
	built    map[string]any
	failed   map[string]error
}

// BuildBehaviorProviders builds each behavior's provider for host's platform
// once, at startup, and logs the outcome. A behavior with no provider on this
// platform has its methods skipped; one whose provider failed to build has
// them fail with that error.
func BuildBehaviorProviders(
	h host.Host,
	logger *slog.Logger,
) *BehaviorProviders {
	logger = logger.With(slog.String("subsystem", "agent.behavior"))
	if h.Logger == nil {
		h.Logger = logger
	}

	bp := &BehaviorProviders{
		platform: h.Platform,
		built:    map[string]any{},
		failed:   map[string]error{},
	}
	// One builder, so that a behavior that others require is built once,
	// and each of them gets the same provider.
	builder := host.NewBuilder(h)
	for _, b := range behavior.All() {
		p, err := builder.Build(b)
		switch {
		case errors.Is(err, provider.ErrUnsupported) || (err == nil && p == nil):
			logger.Info("behavior unsupported on this platform; its methods are skipped",
				slog.String("behavior", b.Name), slog.String("platform", string(h.Platform)))
		case err != nil:
			bp.failed[b.Name] = err
			logger.Error("behavior provider failed to build",
				slog.String("behavior", b.Name), slog.String("error", err.Error()))
		default:
			bp.built[b.Name] = p
			logger.Debug("behavior provider built",
				slog.String("behavior", b.Name), slog.String("platform", string(h.Platform)))
		}
	}

	return bp
}

// NewBehaviorProcessor returns a ProcessorFunc that runs any registered
// behavior's method with the provider built for it, so a behavior needs no
// wiring in the agent.
func NewBehaviorProcessor(
	providers *BehaviorProviders,
	logger *slog.Logger,
) ProcessorFunc {
	logger = logger.With(slog.String("subsystem", "agent.processor.behavior"))

	return func(ctx context.Context, req job.Request) (json.RawMessage, error) {
		name, method, ok := strings.Cut(req.Operation, ".")
		if !ok {
			return nil, fmt.Errorf("invalid behavior operation: %s", req.Operation)
		}

		b := behavior.Lookup(name)
		if b == nil {
			return nil, fmt.Errorf("unknown behavior: %s", name)
		}

		m := b.Method(method)
		if m == nil {
			return nil, fmt.Errorf("unknown method %s on behavior %s", method, name)
		}

		if err := providers.failed[name]; err != nil {
			return nil, err
		}
		p, ok := providers.built[name]
		if !ok || !b.RunsOnHost(m, providers.platform) {
			return nil, fmt.Errorf(
				"%s on %q: %w",
				name,
				providers.platform,
				provider.ErrUnsupported,
			)
		}

		logger.Debug(
			"executing behavior",
			slog.String("operation", b.Operation(m)),
		)

		data := req.Data
		if len(req.SealedKeys) > 0 {
			opened, err := openSecrets(ctx, m, data)
			if err != nil {
				return nil, err
			}
			data = opened
		}

		// The API checked the input before it queued the job, but a job can
		// reach NATS another way, so the agent checks it again.
		if err := m.Check(ctx, data); err != nil {
			return nil, err
		}

		out, err := m.Invoke(ctx, p, data)
		if err != nil {
			return nil, err
		}

		return json.Marshal(out)
	}
}

// openSecrets is data, the input of m, with each sealed secret opened with
// the Opener of the job. The values live only in the input of this one
// call: the agent writes them to no log, state or result.
func openSecrets(
	ctx context.Context,
	m *behavior.Method,
	data json.RawMessage,
) (json.RawMessage, error) {
	o := seal.OpenerFrom(ctx)
	if o == nil {
		return nil, seal.ErrNotSealedToAgent
	}
	in := reflect.New(m.In)
	if err := json.Unmarshal(data, in.Interface()); err != nil {
		return nil, fmt.Errorf("%w: %s", behavior.ErrInvalid, err)
	}
	if err := secret.Map(in.Interface(), o.Value); err != nil {
		return nil, err
	}

	return json.Marshal(secret.Plain(in.Elem().Interface()))
}
