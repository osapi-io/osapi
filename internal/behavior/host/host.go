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
	"io"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/provider"

	"github.com/avfs/avfs"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/exec"
)

// ErrCommandDisabled is the refusal of a command that a caller supplies, such
// as command exec or docker exec, on an agent where agent.command.enabled is
// false: Host.CommandEnabled.
var ErrCommandDisabled = errors.New(
	"command execution is disabled on this agent: set agent.command.enabled to true to allow it",
)

// Host is one agent's host, as a provider builder sees it.
type Host struct {
	// Platform is the agent's platform as gohai's family.Detect names it, such as
	// debian or darwin.
	Platform behavior.PlatformName
	// Container is true when the agent runs in a container.
	Container bool
	// CommandEnabled is agent.command.enabled: whether the agent runs a
	// command that a caller supplies. A method that runs one, such as
	// docker exec, refuses when it is false.
	CommandEnabled bool
	// Hostname is the agent's hostname.
	Hostname string
	Logger   *slog.Logger
	// FS is the filesystem a provider reads and writes through.
	FS avfs.VFS
	// Exec runs commands, with sudo for RunPrivilegedCmd when enabled.
	Exec Exec
	// FileState records the files osapi deployed, in the bucket
	// FileStateBucket. It is nil when the agent has no store for it.
	FileState provider.FileState
	// FileStateBucket is the KV bucket of FileState.
	FileStateBucket string
	// Objects reads the objects that osapi stores, such as a file uploaded
	// with osapi client file upload. It is nil when the agent has no object
	// store.
	Objects Objects
	// Settings is the behavior's decoded settings, as declared with
	// behavior.Config, or its defaults before the config is read.
	Settings any
	// Required are the providers of the behaviors this one requires, by
	// behavior name, built for the same host. Require reads one.
	Required map[string]any
	// Facts are the agent's facts, by their JSON names, for a template that a
	// provider renders with behavior.Render. The agent gives the facts it
	// collected last; a test gives none, or fixed ones. Nil is no facts.
	Facts func() map[string]any
	// Now is the host's clock. The agent gives time.Now; a test gives a
	// fixed time, so an example that reads the time, such as an uptime from
	// the boot time, has one answer. Time reads it.
	Now func() time.Time
	// Resolver resolves names for gohai, such as the canonical FQDN. The
	// agent gives none, which is the system resolver; a test gives one that
	// answers without the network.
	Resolver source.Resolver
	// Reach is a round trip from the agent to NATS, and nil when it answers.
	// A change that can cut the host off its network, such as netplan's,
	// keeps the change only when Reach answers after it. The agent gives a
	// request to its NATS server; a test gives none, which is a host whose
	// change nothing checks.
	Reach func(ctx context.Context) error
}

// Sources are the sources a gohai collector reads this host through: its
// filesystem, its commands, its clock, its platform as the family, and its
// resolver. A nil field is the real host's.
func (h Host) Sources() source.Sources {
	src := source.Sources{FS: h.FS, Now: h.Now, Family: string(h.Platform), Resolver: h.Resolver}
	if h.Exec != nil {
		src.Exec = runner{exec: h.Exec}
	}

	return src
}

// runner is the host's commands as gohai's source.Runner: each runs as the
// agent's own user, with RunReadCmd, which runs it as gohai's own runner
// does.
type runner struct {
	exec Exec
}

// Execute runs name with args, and is its output. The context says that
// gohai runs it, which FromGohai reads.
func (r runner) Execute(
	ctx context.Context,
	name string,
	args ...string,
) ([]byte, error) {
	return r.exec.RunReadCmd(context.WithValue(ctx, gohaiKey{}, true), name, args)
}

// gohaiKey marks the context of a command that gohai runs.
type gohaiKey struct{}

// FromGohai reports whether gohai runs the command of ctx, through Sources.
// The tests of a behavior read it: a read of a host fact goes through gohai.
func FromGohai(
	ctx context.Context,
) bool {
	v, _ := ctx.Value(gohaiKey{}).(bool)

	return v
}

// Time is the time on the host: h.Now, or time.Now when the host gives none.
func (h Host) Time() time.Time {
	if h.Now == nil {
		return time.Now()
	}

	return h.Now()
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
	// RunReadCmd runs name with args as the agent's own user to read the
	// host, as gohai runs a command: stdout is the data and stderr is in the
	// error, with LC_ALL=C, a timeout and a limit on the output. gohai's
	// runner of the host is this.
	RunReadCmd(
		ctx context.Context,
		name string,
		args []string,
	) ([]byte, error)
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
	// RunPrivilegedCmdInteractive runs name with args, through sudo when
	// privilege escalation is on, and calls interact with its standard input
	// while it runs, for a command that waits for an answer that the caller
	// gives later, such as the confirmation of netplan try. The standard
	// input closes when interact returns. The result is the combined output.
	RunPrivilegedCmdInteractive(
		ctx context.Context,
		name string,
		args []string,
		interact func(ctx context.Context, stdin io.Writer) error,
	) (string, error)
	// RunPrivilegedCmdFull is RunPrivilegedCmd with the standard output,
	// the standard error and the exit code apart, for a command whose
	// caller gets all three, such as docker exec. A command that exits
	// with a code other than 0 is not an error: err is a failure to run it.
	// stdin, when it is not empty, goes to the program's standard input, as
	// RunPrivilegedCmdWithStdin gives it.
	RunPrivilegedCmdFull(
		ctx context.Context,
		name string,
		args []string,
		stdin string,
	) (stdout string, stderr string, exitCode int, err error)
	// RunCmdFull runs name with args as the agent's own user, in cwd when it
	// is not empty, for at most timeout seconds, and returns its standard
	// output, standard error, exit code and duration apart. A program that
	// exits non-zero is a result, not an error.
	RunCmdFull(
		ctx context.Context,
		name string,
		args []string,
		cwd string,
		timeout int,
	) (*exec.CmdResult, error)
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
// for the platform. The provider tests build theirs with it on a fake host.
// It is one build: the agent, which builds every behavior, uses a Builder,
// so that a behavior that several others require is built once.
func Build(
	b *behavior.Behavior,
	h Host,
) (any, error) {
	return NewBuilder(h).Build(b)
}

// Builder builds the providers of one host, each behavior once. A behavior
// that several others require, such as managedfile, gives each of them, and
// the agent, the same provider. Thus a lock or a cache that the provider
// keeps is shared by every job that reaches it. A Builder is not safe for
// concurrent use: the agent builds its providers at startup, one at a time.
type Builder struct {
	host  Host
	built map[string]built
}

// built is the result of one behavior's build.
type built struct {
	provider any
	err      error
}

// NewBuilder is a Builder for h.
func NewBuilder(
	h Host,
) *Builder {
	return &Builder{host: h, built: map[string]built{}}
}

// Build is Build of b on the Builder's host. A behavior that was built
// already gives the provider, or the error, of that build.
func (bd *Builder) Build(
	b *behavior.Behavior,
) (any, error) {
	if r, ok := bd.built[b.Name]; ok {
		return r.provider, r.err
	}

	required := make(map[string]any, len(b.Requires))
	for _, r := range b.Requires {
		p, err := bd.Build(r)
		if err != nil && !errors.Is(err, behavior.ErrUnsupported) {
			return nil, fmt.Errorf("%s requires %s: %w", b.Name, r.Name, err)
		}
		if p != nil {
			required[r.Name] = p
		}
	}

	h := bd.host
	h.Settings = b.CurrentSettings()
	h.Required = required

	p, err := b.Provider(h.Platform, h)
	bd.built[b.Name] = built{provider: p, err: err}

	return p, err
}

// Objects reads the objects of the object store of the fleet.
type Objects interface {
	// Object is the content of the object name, and the content type that
	// its uploader gave it: raw, or template. An object that is not there
	// is an error that wraps behavior.ErrNotFound.
	Object(
		ctx context.Context,
		name string,
	) (content []byte, contentType types.ContentType, err error)
}
