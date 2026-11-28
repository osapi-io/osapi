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

package service

import (
	"context"
	"fmt"
	"strings"

	gohaisvc "github.com/osapi-io/gohai/pkg/gohai/collectors/services"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// errReadOnly is what a platform with no writer returns for a write.
var errReadOnly = fmt.Errorf("services on this platform: %w", behavior.ErrUnsupported)

// reader reads the services through gohai, from systemctl. It does not
// write: a platform with no writer reports skipped.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Services is every service that systemd has loaded.
func (r *reader) Services(
	ctx context.Context,
) ([]Entry, error) {
	out, err := gohaisvc.NewWith(r.src).Collect(ctx, nil)
	list := []Entry{}
	if info, ok := out.(*gohaisvc.Info); ok {
		for _, s := range info.Services {
			list = append(list, Entry{
				Name:        s.Name,
				State:       s.State,
				Enabled:     s.Enabled,
				Description: s.Description,
			})
		}
	}

	return list, err
}

// Service is the service name, with .service or without it.
func (r *reader) Service(
	ctx context.Context,
	name string,
) (Entry, error) {
	name = strings.TrimSuffix(name, ".service")
	list, err := r.Services(ctx)
	for _, s := range list {
		if s.Name == name {
			return s, err
		}
	}
	if err == nil {
		err = fmt.Errorf("service %q: %w", name, behavior.ErrNotFound)
	}

	return Entry{}, err
}

// Deploy cannot write here.
func (r *reader) Deploy(
	context.Context,
	Unit,
	bool,
) (bool, error) {
	return false, errReadOnly
}

// Remove cannot write here.
func (r *reader) Remove(
	context.Context,
	string,
) (bool, error) {
	return false, errReadOnly
}

// Act cannot write here.
func (r *reader) Act(
	context.Context,
	string,
	string,
) (bool, error) {
	return false, errReadOnly
}

// state is the service name as systemd has it: its sub-state, and whether it
// is enabled. A unit that systemd has not loaded is dead and disabled. gohai
// reports a systemctl that fails as no services, so this reads no error.
func (r *reader) state(
	ctx context.Context,
	name string,
) Entry {
	s, err := r.Service(ctx, name)
	if err != nil {
		return Entry{Name: name, State: "dead"}
	}

	return s
}
