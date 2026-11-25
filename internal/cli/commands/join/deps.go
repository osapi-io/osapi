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

package join

import (
	"context"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
)

// Deps is what the join token commands use: the config, what connects to
// NATS as the controller does, with a function that closes the connection,
// and the time.
type Deps struct {
	Config config.Config
	Dial   func(ctx context.Context) (jointoken.KV, func(), error)
	Now    func() time.Time
	// Audit writes an entry to the audit log of the controller, or is nil
	// when the controller keeps none.
	Audit func(ctx context.Context, e audit.Entry) error
}
