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

// Package host is what the agent hands a behavior's provider builder: the
// platform it runs on and the shared things a provider reaches the host
// through. It is apart from package behavior so that the SDK, which imports
// behavior, does not pull in what only the agent needs.
package host

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/avfs/avfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/behavior"
)

// Host is one agent's host, as a provider builder sees it.
type Host struct {
	// Platform is the agent's platform as platform.Detect names it, such as
	// debian or darwin.
	Platform string
	// Container is true when the agent runs in a container.
	Container bool
	// Hostname is the agent's hostname.
	Hostname string
	Logger   *slog.Logger
	// FS is the filesystem a provider reads and writes through.
	FS avfs.VFS
	// Exec runs commands, with sudo for RunPrivilegedCmd when enabled.
	Exec Exec
	// FileState records the files osapi deployed. It is nil when the agent
	// has no store for it.
	FileState jetstream.KeyValue
	// Settings is the behavior's decoded settings, as declared with
	// behavior.Config, or its defaults before the config is read.
	Settings any
	// Required are the providers of the behaviors this one requires, by
	// behavior name, built for the same host. Require reads one.
	Required map[string]any
}

// Exec is what a provider needs of the agent's command runner. It is
// declared here, where providers use it, rather than taken from the agent's
// own package, so a behavior names no type it cannot import and a test can
// fake it.
type Exec interface {
	// RunCmd runs name with args as the agent's own user.
	RunCmd(
		ctx context.Context,
		name string,
		args []string,
	) (string, error)
	// RunPrivilegedCmd runs name with args, through sudo when the agent's
	// privilege escalation is on. A behavior declares each program it runs
	// this way with behavior.Sudo.
	RunPrivilegedCmd(
		ctx context.Context,
		name string,
		args []string,
	) (string, error)
	// RunPrivilegedCmdWithStdin is RunPrivilegedCmd with stdin written to
	// the program's standard input: how a secret reaches a program, since
	// one in its arguments is visible to every user on the host.
	RunPrivilegedCmdWithStdin(
		ctx context.Context,
		name string,
		args []string,
		stdin string,
	) (string, error)
}

// Require is the provider of a behavior that h's behavior declared with
// behavior.Requires, as P, its provider type. It is the zero value when that
// behavior has no provider on this platform, which a builder answers with
// behavior.ErrUnsupported.
func Require[P any](
	h Host,
	b *behavior.Behavior,
) P {
	p, _ := h.Required[b.Name].(P)

	return p
}

// Settings is h's settings as T, the struct the behavior passed to
// behavior.Config. Inside a provider builder the behavior's own variable
// cannot be named, so this is how a builder reads them.
func Settings[T any](
	h Host,
) T {
	v, _ := h.Settings.(T)

	return v
}

// Build builds b's provider for h's platform, with b's settings in the host
// its builder gets, after building the provider of each behavior b requires
// for the same host. It returns nil and no error when b declares no provider
// for the platform. The agent builds every behavior's provider with it at
// startup, and the provider tests build theirs with it on a fake host.
func Build(
	b *behavior.Behavior,
	h Host,
) (any, error) {
	required := make(map[string]any, len(b.Requires))
	for _, r := range b.Requires {
		p, err := Build(r, h)
		if err != nil && !errors.Is(err, behavior.ErrUnsupported) {
			return nil, fmt.Errorf("%s requires %s: %w", b.Name, r.Name, err)
		}
		if p != nil {
			required[r.Name] = p
		}
	}

	h.Settings = b.CurrentSettings()
	h.Required = required

	return b.Provider(h.Platform, h)
}
