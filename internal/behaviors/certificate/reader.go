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

package certificate

import (
	"context"
	"fmt"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/cacertificates"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the certificate authorities through gohai's ca_certificates
// collector: each family's own directories on Linux, and the system
// keychains on macOS. It changes nothing.
type reader struct {
	src source.Sources
}

// newReader is the provider on macOS, where osapi does not change the trust
// store.
func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// List reads every certificate authority that the host trusts, in the order
// that gohai reads them: the system ones of a directory, then the local ones.
func (r *reader) List(
	ctx context.Context,
) ([]Entry, error) {
	out, err := cacertificates.NewWith(r.src).Collect(ctx, nil)
	list := []Entry{}
	if info, ok := out.(*cacertificates.Info); ok {
		for _, c := range info.Certificates {
			list = append(list, Entry{
				Name:        c.Name,
				Path:        c.Path,
				Source:      c.Source,
				Enabled:     c.Enabled,
				Subject:     c.Subject,
				Fingerprint: c.Fingerprint,
				NotAfter:    c.NotAfter,
			})
		}
	}

	return list, err
}

// Get reads one certificate authority by its name.
func (r *reader) Get(
	ctx context.Context,
	name string,
) (Entry, error) {
	list, err := r.List(ctx)

	return find(list, name, err)
}

// Deploy is not supported where osapi does not change the trust store.
func (r *reader) Deploy(
	context.Context,
	string,
	string,
	bool,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Remove is not supported where osapi does not change the trust store.
func (r *reader) Remove(
	context.Context,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// find is the entry of list with the name, or not found.
func find(
	list []Entry,
	name string,
	err error,
) (Entry, error) {
	for _, e := range list {
		if e.Name == name {
			return e, err
		}
	}
	if err == nil {
		err = fmt.Errorf("certificate %q: %w", name, behavior.ErrNotFound)
	}

	return Entry{}, err
}
