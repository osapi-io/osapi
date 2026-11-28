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

package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/avfs/avfs"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsauth"

	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/bus"
)

// SetHostnameFn replaces the function that returns the name of the host.
func SetHostnameFn(
	fn func() (string, error),
) {
	hostnameFn = fn
}

// ResetHostnameFn puts back the function that returns the name of the host.
func ResetHostnameFn() {
	hostnameFn = os.Hostname
}

// SetDialFn replaces the function that connects to NATS.
func SetDialFn(
	fn func(context.Context, *slog.Logger, *natsclient.Options) (*natsclient.Client, error),
) {
	dialFn = fn
}

// ResetDialFn puts back the function that connects to NATS.
func ResetDialFn() {
	dialFn = bus.Dial
}

// SetNatsauthOperator swaps how the server makes its operator.
func SetNatsauthOperator(
	fn func(context.Context, avfs.VFS, string, string, config.NATS) (natsauth.Result, error),
) func() {
	was := natsauthOperator
	natsauthOperator = fn

	return func() { natsauthOperator = was }
}

// SetNatsauthUpdateOperator swaps how the server signs its operator again.
func SetNatsauthUpdateOperator(
	fn func(context.Context, avfs.VFS, string, string, config.NATS) (natsauth.Result, error),
) func() {
	was := natsauthUpdateOperator
	natsauthUpdateOperator = fn

	return func() { natsauthUpdateOperator = was }
}

// ExportWait exposes wait, on a Server that logs to log, for testing.
func ExportWait(
	ctx context.Context,
	log *slog.Logger,
	done <-chan struct{},
	what string,
) {
	(&Server{log: log}).wait(ctx, done, what)
}
