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
	"log/slog"

	"github.com/osapi-io/osapi/internal/behavior/host"

	"github.com/osapi-io/osapi/internal/provider"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
)

// hostEnv is what the providers of a host are made from. Each method returns
// the provider of one area for platform: "debian", "darwin", or any other
// value for the Linux fallback, which does nothing.
type hostEnv struct {
	log       *slog.Logger
	fs        avfs.VFS
	exec      exec.Manager
	platform  string
	container bool
	hostname  string
	// fileState and fileStateBucket record the files that the providers
	// deploy. fileState is nil when the agent has no file-state bucket.
	fileState       provider.FileState
	fileStateBucket string
	// objects are the objects of the object store, for behaviors; nil
	// without an object store.
	objects host.Objects
	// reach is a round trip to the NATS server, for a change that can cut
	// the host off, such as netplan's: nil when the server answers.
	reach func(ctx context.Context) error
}

func (e hostEnv) host() nodeHost.Provider {
	return nodeHost.New()
}

func (e hostEnv) disk() disk.Provider {
	return disk.New()
}

func (e hostEnv) memory() memory.Provider {
	return memory.New()
}

func (e hostEnv) load() load.Provider {
	return load.New()
}
