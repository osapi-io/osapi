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

package route

import (
	"context"
	"fmt"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the routes through gohai: ip route on Linux, netstat on
// macOS. It does not write them: a platform with no writer reports skipped.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Routes is the routes of one interface. An interface the host does not
// have is not found; one with no routes has none.
func (r *reader) Routes(
	ctx context.Context,
	name string,
) ([]Entry, error) {
	info, err := r.read(ctx)
	if err != nil {
		return nil, err
	}
	if !has(info, name) {
		return nil, fmt.Errorf("interface %q: %w", name, behavior.ErrNotFound)
	}

	return entries(info, func(rt network.Route) bool { return rt.Interface == name }), nil
}

// Set is unsupported where no writer is declared.
func (r *reader) Set(
	context.Context,
	Route,
	bool,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Remove is unsupported where no writer is declared.
func (r *reader) Remove(
	context.Context,
	string,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// read is gohai's network facts.
func (r *reader) read(
	ctx context.Context,
) (*network.Info, error) {
	out, err := network.NewWith(r.src).Collect(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("read the routes: %w", err)
	}

	return out.(*network.Info), nil
}

// has reports whether the host has the interface.
func has(
	info *network.Info,
	name string,
) bool {
	for _, i := range info.Interfaces {
		if i.Name == name {
			return true
		}
	}

	return false
}

// entries is the routes of info that keep accepts.
func entries(
	info *network.Info,
	keep func(network.Route) bool,
) []Entry {
	out := []Entry{}
	for _, rt := range info.Routes {
		if !keep(rt) {
			continue
		}
		out = append(out, Entry{
			Destination: rt.Destination,
			Gateway:     rt.Gateway,
			Interface:   rt.Interface,
			Family:      rt.Family,
			Metric:      rt.Metric,
			Proto:       rt.Proto,
			Scope:       rt.Scope,
		})
	}

	return out
}
