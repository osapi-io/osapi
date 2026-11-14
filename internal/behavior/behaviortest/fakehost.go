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
// own test calls Run; osapi behavior new writes that line.
package behaviortest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

// Hostname is the fake host's name.
const Hostname = "test-host"

// errFault is the error a fault returns.
var errFault = errors.New("injected fault")

// errUnsupported is what a call the fake host has no answer for panics with,
// after the mock controller reports it.
var errUnsupported = errors.New("the fake host does not support that call")

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
	later   []behavior.Given

	// values is the store's content, and revision its last revision.
	values   map[string][]byte
	revision uint64

	// calls counts filesystem, command and store calls; the one numbered
	// failAt fails, counting from 1, and 0 fails none.
	calls  int
	failAt int

	// gets counts store reads; the one numbered corruptAt returns corrupt as
	// the stored value, as a damaged record would read.
	gets      int
	corruptAt int
	corrupt   []byte

	// unsupported are the calls the mocks had no answer for.
	unsupported []string

	// sudo and runs are the programs the behavior and the ones it requires
	// declare with behavior.Sudo and behavior.Runs; undeclared is each
	// program run without its declaration, once.
	sudo       []string
	runs       []string
	undeclared []string
}

func newHost(
	b *behavior.Behavior,
	platform string,
) *fakeHost {
	h := &fakeHost{outputs: map[string]string{}, values: map[string][]byte{}}
	h.declare(b)
	ctrl := gomock.NewController(h)
	h.Host = host.Host{
		Platform:  platform,
		Hostname:  Hostname,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		FS:        h.filesystem(),
		Exec:      h.commands(ctrl),
		FileState: h.store(ctrl),
	}

	return h
}

// declare adds the programs b and every behavior it requires declare.
func (h *fakeHost) declare(
	b *behavior.Behavior,
) {
	h.sudo = append(h.sudo, b.Sudo...)
	h.runs = append(h.runs, b.Runs...)
	for _, r := range b.Requires {
		h.declare(r)
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
		problem = fmt.Sprintf("runs %s with sudo but does not declare it with behavior.Sudo", name)
	case !privileged && !slices.Contains(h.runs, name) && !slices.Contains(h.sudo, name):
		problem = fmt.Sprintf("runs %s but does not declare it with behavior.Runs", name)
	}
	if problem != "" && !slices.Contains(h.undeclared, problem) {
		h.undeclared = append(h.undeclared, problem)
	}
}

// Errorf records a call the mocks had no answer for. With Fatalf it makes
// the host the mock controller's reporter.
func (h *fakeHost) Errorf(
	format string,
	args ...any,
) {
	h.unsupported = append(h.unsupported, fmt.Sprintf(format, args...))
}

// Fatalf records the call and stops the method, which the checks recover.
func (h *fakeHost) Fatalf(
	format string,
	args ...any,
) {
	h.Errorf(format, args...)
	panic(errUnsupported)
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
	_ = vfs.SetFailFunc(func(avfs.VFSBase, avfs.FnVFS, *failfs.FailParam) error {
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

		return out, nil
	}

	m := hostmocks.NewMockExec(ctrl)
	m.EXPECT().RunCmd(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, args []string) (string, error) {
			h.check(name, false)

			return run(ctx, name, args)
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

	return m
}

// store is the generated jetstream.KeyValue mock, holding values in memory.
// It answers what providers use: Get, Put, PutString, Create, Update, Delete
// and Purge. Any other call is reported as unsupported.
func (h *fakeHost) store(
	ctrl *gomock.Controller,
) jetstream.KeyValue {
	put := func(_ context.Context, key string, value []byte) (uint64, error) {
		if err := h.next(); err != nil {
			return 0, err
		}
		h.revision++
		h.values[key] = append([]byte(nil), value...)

		return h.revision, nil
	}
	remove := func(_ context.Context, key string, _ ...jetstream.KVDeleteOpt) error {
		if err := h.next(); err != nil {
			return err
		}
		delete(h.values, key)

		return nil
	}

	kv := jobmocks.NewMockKeyValue(ctrl)
	kv.EXPECT().Get(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, key string) (jetstream.KeyValueEntry, error) {
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

			return h.entry(ctrl, key, value), nil
		}).AnyTimes()
	kv.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(put).AnyTimes()
	kv.EXPECT().PutString(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, value string) (uint64, error) {
			return put(ctx, key, []byte(value))
		}).AnyTimes()
	kv.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			key string,
			value []byte,
			_ ...jetstream.KVCreateOpt,
		) (uint64, error) {
			if _, ok := h.values[key]; ok {
				return 0, jetstream.ErrKeyExists
			}

			return put(ctx, key, value)
		}).AnyTimes()
	kv.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, value []byte, _ uint64) (uint64, error) {
			return put(ctx, key, value)
		}).AnyTimes()
	kv.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(remove).AnyTimes()
	kv.EXPECT().Purge(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(remove).AnyTimes()

	return kv
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
