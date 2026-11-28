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

// Package behaviortest runs a behavior's examples and refusals against a fake
// host, so a behavior's provider is tested from its definition. A behavior's
// own test calls Run; osapi dsl behavior create writes that line.
package behaviortest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	sourcemocks "github.com/osapi-io/gohai/pkg/gohai/source/mocks"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/exec"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

// Now is the fake host's time, for each example: an output that reads the
// clock, such as an uptime, is computed from it.
var Now = time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

// Hostname is the fake host's name.
const Hostname = "test-host"

// errFault is the error a fault returns.
var errFault = errors.New("injected fault")

// fakeHost is one host for one example: an in-memory filesystem, a command
// runner that records each command line and prints what the example's Prints
// says, and an in-memory file state store. The runner and the store are the
// generated mocks, configured here.
type fakeHost struct {
	host.Host

	// ran is every command line run during the example, in order; history
	// is every one run on the host, the example's setup included. outputs
	// is what a command line prints, and later what it prints once another
	// has run, declared with PrintsAfter.
	ran     []string
	history []string
	outputs map[string]string
	// exits are the exit codes of the commands declared with Exits, by
	// command line.
	exits map[string]int
	later []behavior.Given
	// laterFiles are the files that a command writes when it runs,
	// declared with ExistsAfter.
	laterFiles []behavior.Given

	// values is the store's content, and revision its last revision.
	values   map[string][]byte
	revision uint64

	// objects are the objects of the object store, by name.
	objects map[string]behavior.Given

	// calls counts filesystem, command and store calls; the one numbered
	// failAt fails, counting from 1, and 0 fails none.
	calls  int
	failAt int

	// gets counts store reads; the one numbered corruptAt returns corrupt as
	// the stored value, as a damaged record would read.
	gets      int
	corruptAt int
	corrupt   []byte

	// sudo and runs are the programs the behavior and the ones it requires
	// declare with behavior.Sudo and behavior.Runs; undeclared is each
	// program run without its declaration, once.
	sudo       []string
	runs       []string
	undeclared []string

	// writes are the paths the behavior under test declares with
	// behavior.Writes. A required behavior's paths do not count: a behavior
	// declares each path it writes, also through managedfile. keepsState is
	// set when it or a behavior it requires declares behavior.KeepsState.
	// setup is set while prepare writes an example's Given files, which are
	// not the behavior's writes.
	writes     []string
	keepsState bool
	setup      bool

	// read is the read under test that must read through gohai: set while
	// a read with no NotAFact runs.
	read bool
}

func newHost(
	b *behavior.Behavior,
	platform behavior.PlatformName,
) *fakeHost {
	h := &fakeHost{
		outputs: map[string]string{},
		exits:   map[string]int{},
		values:  map[string][]byte{},
		objects: map[string]behavior.Given{},
		writes:  b.Writes,
	}
	h.declare(b, platform)
	ctrl := gomock.NewController(h)
	h.Host = host.Host{
		Platform: platform,
		Hostname: Hostname,
		// The examples run on an agent that runs a caller's command, as the
		// integration test's agent does; docker exec needs it.
		CommandEnabled: true,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		FS:             h.filesystem(),
		Exec:           h.commands(ctrl),
		FileState:      fileState{h: h, ctrl: ctrl},
		Objects:        objects{h: h},
		Now:            func() time.Time { return Now },
		Resolver:       noNames(ctrl),
	}

	return h
}

// place puts g, a file or a link that an example gives, on the host as its
// setup, with its directory. An error is a Given that the host cannot hold,
// such as a file under a path that is a file: the example would run on an
// empty host and pass or fail for the wrong reason.
func (h *fakeHost) place(
	g behavior.Given,
) error {
	h.setup = true
	defer func() { h.setup = false }()
	if err := h.FS.MkdirAll(filepath.Dir(g.Path), 0o755); err != nil {
		return err
	}
	if g.Kind == behavior.GivenLink {
		return h.FS.Symlink(g.Content, g.Path)
	}

	return h.FS.WriteFile(g.Path, []byte(g.Content), 0o644)
}

// declare adds the programs b declares for the methods platform's provider
// serves, and every program of each behavior b requires. An empty platform
// is every method.
func (h *fakeHost) declare(
	b *behavior.Behavior,
	platform behavior.PlatformName,
) {
	h.sudo = append(append(h.sudo, b.Sudo...), b.ProviderSudoOn(platform)...)
	_, runs := b.Programs(platform)
	h.runs = append(append(h.runs, runs...), b.RunsIfPresent...)
	for _, m := range b.Methods {
		if platform == "" || b.Serves(m, platform) {
			h.sudo = append(h.sudo, m.SudoPrograms...)
			h.runs = append(h.runs, m.RunsPrograms...)
		}
	}
	h.keepsState = h.keepsState || b.KeepsState
	for _, r := range b.Requires {
		h.declare(r, "")
	}
}

// check records a program run without the declaration it needs: Sudo for
// one run with sudo, Runs or Sudo for one run without.
func (h *fakeHost) check(
	name string,
	privileged bool,
) {
	var problem string
	switch {
	case privileged && !slices.Contains(h.sudo, name):
		problem = behavior.TestFailure(behavior.TestRuns,
			fmt.Sprintf("runs %s with sudo but does not declare it with behavior.Sudo", name))
	case !privileged && !slices.Contains(h.runs, name) && !slices.Contains(h.sudo, name) &&
		!slices.Contains(h.runs, behavior.AnyProgram):
		problem = behavior.TestFailure(behavior.TestRuns,
			fmt.Sprintf("runs %s but does not declare it with behavior.Runs", name))
	}
	h.report(problem)
}

// report records problem once. An empty problem is none.
func (h *fakeHost) report(
	problem string,
) {
	if problem != "" && !slices.Contains(h.undeclared, problem) {
		h.undeclared = append(h.undeclared, problem)
	}
}

// tempSuffix is the suffix atomicfile.Write gives the temporary file beside
// its target, which is a write of the target.
var tempSuffix = regexp.MustCompile(`\.osapi-\d+-\d+$`)

// checkWrite records a change to p, a write, a removal or a mode change,
// that is not under a path the behavior declares with behavior.Writes.
func (h *fakeHost) checkWrite(
	p string,
) {
	if h.setup {
		return
	}
	p = tempSuffix.ReplaceAllString(p, "")
	if !fromAtomicfile() {
		h.report(behavior.TestFailure(behavior.TestFileSpan, fmt.Sprintf(
			"changes %s through h.FS, so the trace has no file span for it", p)))
	}
	for _, pattern := range h.writes {
		if ok, _ := path.Match(pattern, p); ok || pattern == "*" {
			return
		}
	}
	h.report(behavior.TestFailure(behavior.TestWrites,
		fmt.Sprintf("writes %s but does not declare it with behavior.Writes", p)))
}

// fromAtomicfile reports whether the filesystem call in progress comes from
// the atomicfile package, which opens the file.write, file.remove and
// file.chmod spans. A provider that changes a file through h.FS directly
// makes no span.
func fromAtomicfile() bool {
	pcs := make([]uintptr, 64)
	frames := runtime.CallersFrames(pcs[:runtime.Callers(3, pcs)])
	for {
		f, more := frames.Next()
		if strings.Contains(f.Function, "/internal/atomicfile.") {
			return true
		}
		if !more {
			return false
		}
	}
}

// checkState records a use of the file-state bucket by a behavior that does
// not declare behavior.KeepsState.
func (h *fakeHost) checkState() {
	if !h.keepsState {
		h.report(behavior.TestFailure(behavior.TestKeepsState,
			"keeps state in the file-state bucket but does not declare behavior.KeepsState"))
	}
}

// Errorf is the reporter of the mock controller. The command mock answers
// each method of host.Exec, so a report is a fault of the fake host itself.
// It joins the problems that the check of the behavior reports.
func (h *fakeHost) Errorf(
	format string,
	args ...any,
) {
	h.report(fmt.Sprintf("the fake host has no answer for a call: "+format, args...))
}

// Fatalf is Errorf, for the mock controller.
func (h *fakeHost) Fatalf(
	format string,
	args ...any,
) {
	h.Errorf(format, args...)
}

// next counts a call and fails the one numbered failAt.
func (h *fakeHost) next() error {
	h.calls++
	if h.calls == h.failAt {
		return errFault
	}

	return nil
}

// filesystem is an in-memory filesystem whose calls count, and fail, with the
// commands and the store's.
func (h *fakeHost) filesystem() avfs.VFS {
	vfs := failfs.New(memfs.New())
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, fn avfs.FnVFS, fp *failfs.FailParam) error {
		switch fn {
		case avfs.FnOpenFile:
			if fp.Flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND) != 0 {
				h.checkWrite(fp.Path)
			}
		case avfs.FnWriteFile, avfs.FnRemove, avfs.FnRemoveAll, avfs.FnChmod, avfs.FnTruncate,
			avfs.FnChown, avfs.FnLchown, avfs.FnChtimes:
			h.checkWrite(fp.Path)
		case avfs.FnRename, avfs.FnSymlink, avfs.FnLink:
			h.checkWrite(fp.NewPath)
		}
		// MkdirAll, Mkdir and CreateTemp are not checked: a parent directory
		// is not a file that behavior.Writes names, and a temp file is made
		// for the rename onto the path, which is checked.

		return h.next()
	})

	return vfs
}

// commands is the generated host.Exec mock, answering every method from
// outputs and recording each command line.
func (h *fakeHost) commands(
	ctrl *gomock.Controller,
) host.Exec {
	run := func(_ context.Context, name string, args []string) (string, error) {
		line := strings.Join(append([]string{name}, args...), " ")
		if err := h.next(); err != nil {
			return "", err
		}
		out := h.outputs[line]
		for _, g := range h.later {
			if g.Command == line && slices.Contains(h.history, g.After) {
				out = g.Content
			}
		}
		h.ran = append(h.ran, line)
		h.history = append(h.history, line)
		for _, g := range h.laterFiles {
			if g.After != line {
				continue
			}
			if err := h.place(g); err != nil {
				h.undeclared = append(h.undeclared, fmt.Sprintf(
					"the example's host cannot hold %s after %s: %v", g.Path, line, err))
			}
		}

		return out, nil
	}

	m := hostmocks.NewMockExec(ctrl)
	m.EXPECT().RunCmd(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string) (string, error) {
			h.check(name, false)
			if h.read && !h.setup && !host.FromGohai(ctx) {
				h.report(behavior.TestFailure(behavior.TestGohai,
					fmt.Sprintf("a read runs %s outside gohai", name)))
			}

			return run(ctx, name, args)
		}).
		AnyTimes()
	// gohai's reads run through RunReadCmd, as the agent runs them.
	m.EXPECT().RunReadCmd(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string) ([]byte, error) {
			h.check(name, false)
			out, err := run(ctx, name, args)

			return []byte(out), err
		}).
		AnyTimes()
	m.EXPECT().RunCmdFull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string, _ string, _ int) (*exec.CmdResult, error) {
			h.check(name, false)
			out, err := run(ctx, name, args)
			if err != nil {
				return nil, err
			}

			line := strings.Join(append([]string{name}, args...), " ")

			return &exec.CmdResult{Stdout: out, ExitCode: h.exits[line]}, nil
		}).
		AnyTimes()
	m.EXPECT().RunPrivilegedCmd(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string) (string, error) {
			h.check(name, true)

			return run(ctx, name, args)
		}).
		AnyTimes()
	m.EXPECT().RunPrivilegedCmdWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string, _ string) (string, error) {
			h.check(name, true)

			return run(ctx, name, args)
		}).
		AnyTimes()

	m.EXPECT().RunPrivilegedCmdFull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string, _ string) (string, string, int, error) {
			h.check(name, true)
			out, err := run(ctx, name, args)
			line := strings.Join(append([]string{name}, args...), " ")

			return out, "", h.exits[line], err
		}).
		AnyTimes()

	return m
}

// fileState is the file-state store of the fake host, holding values in
// memory. Each call can fail in turn, and a read can return a damaged record,
// as next and corruptAt say. The bucket is not used: the host has one.
type fileState struct {
	h    *fakeHost
	ctrl *gomock.Controller
}

// KVGet returns the value of key, or jetstream.ErrKeyNotFound.
func (s fileState) KVGet(
	_ context.Context,
	_ string,
	key string,
) (jetstream.KeyValueEntry, error) {
	h := s.h
	h.checkState()
	if err := h.next(); err != nil {
		return nil, err
	}
	h.gets++
	value, ok := h.values[key]
	if h.gets == h.corruptAt {
		value, ok = h.corrupt, true
	}
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}

	return h.entry(s.ctrl, key, value), nil
}

// KVPut stores value at key, and returns the new revision.
func (s fileState) KVPut(
	_ context.Context,
	_ string,
	key string,
	value []byte,
) (uint64, error) {
	h := s.h
	h.checkState()
	if err := h.next(); err != nil {
		return 0, err
	}
	h.revision++
	h.values[key] = append([]byte(nil), value...)

	return h.revision, nil
}

// entry is the generated jetstream.KeyValueEntry mock for one stored value.
func (h *fakeHost) entry(
	ctrl *gomock.Controller,
	key string,
	value []byte,
) jetstream.KeyValueEntry {
	e := jobmocks.NewMockKeyValueEntry(ctrl)
	e.EXPECT().Key().Return(key).AnyTimes()
	e.EXPECT().Value().Return(value).AnyTimes()
	e.EXPECT().Revision().Return(h.revision).AnyTimes()
	e.EXPECT().Created().Return(time.Time{}).AnyTimes()

	return e
}

// noNames is gohai's generated resolver, answering that no name resolves,
// so an example makes no DNS lookup.
func noNames(
	ctrl *gomock.Controller,
) *sourcemocks.MockResolver {
	m := sourcemocks.NewMockResolver(ctrl)
	notFound := &net.DNSError{Err: "no such host", IsNotFound: true}
	m.EXPECT().LookupHost(gomock.Any(), gomock.Any()).Return(nil, notFound).AnyTimes()
	m.EXPECT().LookupAddr(gomock.Any(), gomock.Any()).Return(nil, notFound).AnyTimes()

	return m
}

// objects is the object store of the fake host. A read counts, and fails,
// with the filesystem, command and store calls.
type objects struct {
	h *fakeHost
}

// Object is the object name that the example stores with Stored.
func (o objects) Object(
	_ context.Context,
	name string,
) ([]byte, types.ContentType, error) {
	if err := o.h.next(); err != nil {
		return nil, "", err
	}
	g, ok := o.h.objects[name]
	if !ok {
		return nil, "", fmt.Errorf("object %q: %w", name, behavior.ErrNotFound)
	}

	return []byte(g.Content), g.ContentType, nil
}
