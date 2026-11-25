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

package natssetup

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/job"
)

// Setup connects as the controller does, with controller.nats, and creates
// or updates each stream and bucket to match the config.
func Setup(
	ctx context.Context,
	d Deps,
	_ Input,
) (Output, error) {
	n := d.Config.Controller.NATS
	job.Init(n.Namespace)

	opts, err := bus.ConnectionOptions(n, "osapi-nats-setup")
	if err != nil {
		return Output{}, err
	}
	if err := d.Prepare(
		ctx,
		d.Logger.With("component", "nats"),
		opts,
		n.Namespace,
		d.Config.NATS,
	); err != nil {
		return Output{}, fmt.Errorf("set up JetStream on %s:%d: %w", n.Host, n.Port, err)
	}

	return Output{Host: n.Host, Port: n.Port, Namespace: n.Namespace, Auth: n.Auth.Type}, nil
}
