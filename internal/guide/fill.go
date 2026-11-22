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
package guide

import "strings"

// Fill is page with the text between each pair of markers of name written
// again by render: <!-- osapi:<name> args --> starts the text, and
// <!-- /osapi:<name> --> ends it. render gets the args, the words after the
// name. A page keeps its own text around the markers. A start with no end is
// left as it is.
func Fill(
	page string,
	name string,
	render func(args []string) string,
) string {
	start := "<!-- osapi:" + name
	end := "<!-- /osapi:" + name + " -->"
	var sb strings.Builder
	for {
		i := strings.Index(page, start)
		if i < 0 {
			break
		}
		head := page[i+len(start):]
		closing := strings.Index(head, "-->")
		stop := strings.Index(head, end)
		if closing < 0 || stop < closing {
			break
		}
		sb.WriteString(page[:i+len(start)+closing+len("-->")])
		// Blank lines around the text, as Prettier writes them.
		sb.WriteString("\n\n" + render(strings.Fields(head[:closing])) + "\n\n")
		page = head[stop:]
	}
	sb.WriteString(page)

	return sb.String()
}
