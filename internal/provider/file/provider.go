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

package file

import (
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"

	"github.com/retr0h/osapi/internal/provider"
)

// Compile-time interface check.
var _ Provider = (*FileProvider)(nil)

// FileProvider implements the Provider interface for file deploy and status
// operations using NATS Object Store for content and KV for state tracking.
type FileProvider struct {
	provider.FactsAware

	logger   *slog.Logger
	fs       afero.Fs
	objStore jetstream.ObjectStore
	stateKV  jetstream.KeyValue
	hostname string
}

// NewFileProvider creates a new FileProvider with the given dependencies.
// Facts are not available at construction time; call SetFactsFunc after
// the agent is initialized to wire template rendering to live facts.
func NewFileProvider(
	logger *slog.Logger,
	fs afero.Fs,
	objStore jetstream.ObjectStore,
	stateKV jetstream.KeyValue,
	hostname string,
) *FileProvider {
	return &FileProvider{
		logger:   logger,
		fs:       fs,
		objStore: objStore,
		stateKV:  stateKV,
		hostname: hostname,
	}
}
