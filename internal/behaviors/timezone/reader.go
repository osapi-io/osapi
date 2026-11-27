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

package timezone

import (
	"context"
	"fmt"

	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/validation"
)

// reader reads the timezone through gohai on every platform, and sets it
// with timedatectl where update runs.
type reader struct {
	src  source.Sources
	exec host.Exec
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources(), exec: h.Exec}, nil
}

// Timezone reads the timezone through gohai.
func (r *reader) Timezone(
	ctx context.Context,
) (Entry, error) {
	return read(ctx, r.src)
}

// SetTimezone sets the timezone with timedatectl set-timezone, unless
// gohai reads that the host already has it.
func (r *reader) SetTimezone(
	ctx context.Context,
	name string,
) (bool, error) {
	// The name becomes an argument of timedatectl: check it here too.
	if len(name) > 64 || !validation.TimezonePattern.MatchString(name) {
		return false, fmt.Errorf("%q is not a valid timezone name", name)
	}
	// gohai's timezone collector returns no error: a zone that it cannot
	// read is empty, and an empty zone never matches a valid name.
	if current, _ := read(ctx, r.src); current.Timezone == name {
		return false, nil
	}
	if _, err := r.exec.RunPrivilegedCmd(
		ctx,
		"timedatectl",
		[]string{"set-timezone", name},
	); err != nil {
		return false, fmt.Errorf("set the timezone: %w", err)
	}

	return true, nil
}
