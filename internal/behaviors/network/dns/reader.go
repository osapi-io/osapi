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

package dns

import (
	"context"
	"fmt"

	gohaidns "github.com/osapi-io/gohai/pkg/gohai/collectors/dns"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the resolvers through gohai: resolvectl and resolv.conf on
// Linux, scutil on macOS. It does not set them: a platform with no writer
// reports skipped.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Resolvers is the interface's own resolvers. A host that has none per
// interface, such as a container with no systemd-resolved, gives its
// resolv.conf for each interface; a host that has some, and none for
// name, has no such interface.
func (r *reader) Resolvers(
	ctx context.Context,
	name string,
) (Entry, error) {
	out, _ := gohaidns.NewWith(r.src).Collect(ctx, nil)
	info := &gohaidns.Info{}
	if read, ok := out.(*gohaidns.Info); ok {
		info = read
	}
	if len(info.Interfaces) == 0 {
		return Entry{Servers: orEmpty(info.Nameservers), SearchDomains: orEmpty(info.Search)}, nil
	}
	res, ok := info.Interfaces[name]
	if !ok {
		return Entry{}, fmt.Errorf("interface %q: %w", name, behavior.ErrNotFound)
	}

	return Entry{Servers: orEmpty(res.Nameservers), SearchDomains: orEmpty(res.Domains)}, nil
}

// Set is unsupported where no writer is declared.
func (r *reader) Set(
	context.Context,
	UpdateInput,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Remove is unsupported where no writer is declared.
func (r *reader) Remove(
	context.Context,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// orEmpty is list, or an empty list for nil, so the JSON is [] not null.
func orEmpty(
	list []string,
) []string {
	if list == nil {
		return []string{}
	}

	return list
}
