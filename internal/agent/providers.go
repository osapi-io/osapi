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
	dockerProv "github.com/osapi-io/osapi/internal/provider/container/docker"
	fileProv "github.com/osapi-io/osapi/internal/provider/file"
	"github.com/osapi-io/osapi/internal/provider/node/disk"
	nodeHost "github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	userProv "github.com/osapi-io/osapi/internal/provider/node/user"
	cronProv "github.com/osapi-io/osapi/internal/provider/schedule/cron"
)

// dockerNewFn makes the Docker client (injectable for testing). docker checks
// the error before it uses the client.
var dockerNewFn = func() (dockerProv.Provider, error) {
	return dockerProv.New()
}

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
	files           fileProv.Provider
	// objects are the objects of the object store, for behaviors; nil
	// without an object store.
	objects host.Objects
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

// docker returns the Docker provider, or nil if Docker does not answer. The
// agent runs without the docker operations then.
func (e hostEnv) docker(
	ctx context.Context,
) dockerProv.Provider {
	c, err := dockerNewFn()
	if err != nil {
		e.log.Info("Docker client creation failed, docker operations disabled",
			slog.String("error", err.Error()))
		return nil
	}
	if err := c.Ping(ctx); err != nil {
		e.log.Info("Docker not available, docker operations disabled",
			slog.String("error", err.Error()))
		return nil
	}

	return c
}

func (e hostEnv) cron() cronProv.Provider {
	switch e.platform {
	case "debian":
		if e.files == nil {
			e.log.Warn("file provider not available, cron operations disabled")
			return cronProv.NewLinuxProvider()
		}
		return cronProv.NewDebianProvider(
			e.log,
			e.fs,
			e.files,
			e.fileState,
			e.fileStateBucket,
			e.hostname,
		)
	case "darwin":
		return cronProv.NewDarwinProvider()
	default:
		return cronProv.NewLinuxProvider()
	}
}

func (e hostEnv) user() userProv.Provider {
	switch e.platform {
	case "debian":
		return userProv.NewDebianProvider(e.log, e.fs, e.exec)
	case "darwin":
		return userProv.NewDarwinProvider()
	default:
		return userProv.NewLinuxProvider()
	}
}
