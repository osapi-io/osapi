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

package controller

import (
	"context"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/notify"
)

// startConditionWatcher watches the registry bucket for condition changes,
// and logs each one, if cfg turns notifications on.
func startConditionWatcher(
	ctx context.Context,
	log *slog.Logger,
	registry notify.Registry,
	registryBucket string,
	cfg config.NotificationsConfig,
) {
	if !cfg.Enabled || registryBucket == "" {
		return
	}

	var renotify time.Duration
	if cfg.RenotifyInterval != "" {
		d, err := time.ParseDuration(cfg.RenotifyInterval)
		if err != nil {
			log.Warn(
				"invalid renotify_interval, using default (0 = disabled)",
				slog.String("value", cfg.RenotifyInterval),
				slog.String("error", err.Error()),
			)
		} else {
			renotify = d
		}
	}

	watcher := notify.NewWatcher(
		registry,
		registryBucket,
		notify.NewLogNotifier(log),
		log,
		renotify,
	)
	background(ctx, log, "condition watcher", watcher.Start)
}
