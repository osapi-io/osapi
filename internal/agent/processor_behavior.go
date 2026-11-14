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
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider"
)

// BehaviorProviders are the providers the agent built for every registered
// behavior on its platform, as each declares them with behavior.Platform.
type BehaviorProviders struct {
	platform string
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
	for _, b := range behavior.All() {
		p, err := host.Build(b, h)
		switch {
		case errors.Is(err, provider.ErrUnsupported) || (err == nil && p == nil):
			logger.Info("behavior unsupported on this platform; its methods are skipped",
				slog.String("behavior", b.Name), slog.String("platform", h.Platform))
		case err != nil:
			bp.failed[b.Name] = err
			logger.Error("behavior provider failed to build",
				slog.String("behavior", b.Name), slog.String("error", err.Error()))
		default:
			bp.built[b.Name] = p
			logger.Debug("behavior provider built",
				slog.String("behavior", b.Name), slog.String("platform", h.Platform))
		}
	}

	return bp
}

// All is every provider built, for the registry, which wires the agent's
// facts into each.
func (bp *BehaviorProviders) All() []any {
	all := make([]any, 0, len(bp.built))
	for _, b := range behavior.All() {
		if p, ok := bp.built[b.Name]; ok {
			all = append(all, p)
		}
	}

	return all
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
		if !ok {
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

		out, err := m.Invoke(ctx, p, req.Data)
		if err != nil {
			return nil, err
		}

		return json.Marshal(out)
	}
}
