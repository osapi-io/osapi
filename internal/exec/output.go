// Copyright (c) 2024 John Dewey

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

package exec

import (
	"bytes"
	"fmt"
)

// output is the writer of a command's output. It keeps the first max bytes
// and counts the rest, so a command that prints without an end, such as
// yes, does not fill the memory of the agent. A write never fails: the
// command runs to its end as it would with no limit.
type output struct {
	buf     bytes.Buffer
	max     int
	dropped int64
}

// Write keeps what fits in o and counts the rest.
func (o *output) Write(
	p []byte,
) (int, error) {
	room := o.max - o.buf.Len()
	if room >= len(p) {
		return o.buf.Write(p)
	}
	if room > 0 {
		o.buf.Write(p[:room])
	}
	o.dropped += int64(len(p) - max(room, 0))

	return len(p), nil
}

// String is the output that o kept, with a last line that counts the bytes
// that it did not keep.
func (o *output) String() string {
	if o.dropped == 0 {
		return o.buf.String()
	}

	return fmt.Sprintf("%s\n[%d more bytes of output not kept]\n", o.buf.String(), o.dropped)
}
