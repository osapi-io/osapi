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

package cmd

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
)

// resolveReferences puts the value of each reference, secret://<vault>/<key>,
// of appConfig under prefix in its place. A command resolves the references
// of the values that it reads itself, from the vaults of its config, as the
// controller does at startup. Thus it connects to NATS as controller.nats
// tells, and only when a value under prefix is a reference.
func resolveReferences(
	ctx context.Context,
	prefix string,
) error {
	if len(config.References(&appConfig, prefix)) == 0 {
		return nil
	}
	n := appConfig.Controller.NATS
	opts, err := bus.ConnectionOptions(n, "osapi-references")
	if err != nil {
		return fmt.Errorf("resolve the references of the config: %w", err)
	}
	nc, err := bus.Dial(logger, opts)
	if err != nil {
		return fmt.Errorf("resolve the references of the config: %w", err)
	}
	defer bus.Close(nc)
	vaults, err := controller.OpenVaults(ctx, nc, appConfig)
	if err != nil {
		return fmt.Errorf("resolve the references of the config: %w", err)
	}
	if err := config.ResolveReferences(ctx, &appConfig, prefix, vaults.Resolve); err != nil {
		return fmt.Errorf("resolve the references of the config: %w", err)
	}

	return nil
}
