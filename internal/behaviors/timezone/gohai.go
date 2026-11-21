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

	gohaitz "github.com/osapi-io/gohai/pkg/gohai/collectors/timezone"
	"github.com/osapi-io/gohai/pkg/gohai/source"
)

// read is the timezone of the host, as gohai reads it from src: the name
// that /etc/localtime links to, and the offset of the host's clock.
func read(
	ctx context.Context,
	src source.Sources,
) (Entry, error) {
	out, err := gohaitz.NewWith(src).Collect(ctx, nil)
	var e Entry
	if info, ok := out.(*gohaitz.Info); ok {
		e = Entry{Timezone: info.Name, UTCOffset: offset(info.Offset)}
	}

	return e, err
}

// offset is seconds east of UTC as +hh:mm.
func offset(
	seconds int,
) string {
	sign := "+"
	if seconds < 0 {
		sign, seconds = "-", -seconds
	}

	return fmt.Sprintf("%s%02d:%02d", sign, seconds/3600, seconds%3600/60)
}
