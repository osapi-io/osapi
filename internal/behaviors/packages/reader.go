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

package packages

import (
	"context"
	"fmt"

	gohaipkg "github.com/osapi-io/gohai/pkg/gohai/collectors/packages"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/validation"
)

// reader reads the installed packages through gohai: dpkg on the Debian
// family, rpm on the other Linux families and Homebrew on macOS. It does
// not install or remove: a platform with no writer reports skipped.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// List is every package that gohai finds installed.
func (r *reader) List(
	ctx context.Context,
) ([]Installed, error) {
	out, err := gohaipkg.NewWith(r.src).Collect(ctx, nil)
	list := []Installed{}
	if info, ok := out.(*gohaipkg.Info); ok {
		for _, p := range info.Packages {
			list = append(list, Installed{
				Name:         p.Name,
				Version:      p.Version,
				Architecture: p.Architecture,
				Manager:      p.PackageManager,
				Description:  p.Description,
				Size:         p.Size,
			})
		}
	}

	return list, err
}

// Get is the installed package named name.
func (r *reader) Get(
	ctx context.Context,
	name string,
) (Installed, error) {
	if err := check(name); err != nil {
		return Installed{}, err
	}
	list, err := r.List(ctx)
	for _, p := range list {
		if p.Name == name {
			return p, err
		}
	}
	if err == nil {
		err = fmt.Errorf("package %q: %w", name, behavior.ErrNotFound)
	}

	return Installed{}, err
}

// Install is unsupported where no writer is declared.
func (r *reader) Install(
	context.Context,
	string,
	string,
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

// check refuses a name that the package manager could read as an option,
// or that is too long: the API checks the same, and this holds for any
// caller.
func check(
	name string,
) error {
	if len(name) > 255 || !validation.PackageNamePattern.MatchString(name) {
		return fmt.Errorf("%q is not a valid package name", name)
	}

	return nil
}
