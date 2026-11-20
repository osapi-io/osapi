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

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
)

// termWidth is the width that PrintTerms wraps the text of a term to.
const termWidth = 76

// PrintTerms prints terms under the current heading: each name in bold, then
// what it does, then when to use it in gray. The text wraps and is indented
// under the name, so a long description is never cut.
func PrintTerms(
	terms []guide.Term,
) {
	FlushKV()
	headed = false
	for i, t := range terms {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println("  " + render(os.Stdout, NameStyle, t.Name))
		for _, line := range wrapWords(t.Does, termWidth-4) {
			fmt.Println("    " + line)
		}
		if t.When != "" {
			for _, line := range wrapWords("Use: "+t.When, termWidth-4) {
				fmt.Println("    " + Mute(os.Stdout, line))
			}
		}
	}
}

// PrintBlock prints one block of a list of blocks under the current
// heading: title in bold, then each pair, a label and a text, wrapped and
// indented under it. A pair with a label prints "Label: text" in gray; a
// pair with none is the text alone. A blank line goes between two blocks,
// and none after the heading.
func PrintBlock(
	title string,
	pairs ...string,
) {
	FlushKV()
	if !headed {
		fmt.Println()
	}
	headed = false
	fmt.Println("  " + render(os.Stdout, NameStyle, title))
	for i := 0; i+1 < len(pairs); i += 2 {
		label, text := pairs[i], pairs[i+1]
		if label == "" {
			for _, line := range wrapWords(text, termWidth-4) {
				fmt.Println("    " + line)
			}

			continue
		}
		for _, line := range wrapWords(label+": "+text, termWidth-4) {
			fmt.Println("    " + Mute(os.Stdout, line))
		}
	}
}

// wrapWords breaks s into lines of at most width characters, at spaces. A
// word longer than width is a line of its own.
func wrapWords(
	s string,
	width int,
) []string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(s) {
		switch {
		case line == "":
			line = w
		case len(line)+1+len(w) > width:
			lines = append(lines, line)
			line = w
		default:
			line += " " + w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}

	return lines
}
