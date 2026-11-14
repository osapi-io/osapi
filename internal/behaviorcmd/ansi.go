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

package behaviorcmd

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

// ansi256 holds the few 256-color codes the theme uses, as hex. Every other
// color in the theme is true color and carries its own RGB.
var ansi256 = map[int]string{
	222: "#ffdf87",
}

// ANSIToHTML turns the theme's output into HTML: each styled run becomes a
// span with an inline color and weight, so a docs page shows the colors the
// terminal does without any script. It understands what the theme emits:
// reset, bold, and 24-bit or 256-color foregrounds.
func ANSIToHTML(
	s string,
) string {
	var (
		sb    strings.Builder
		color string
		bold  bool
		open  bool
	)

	closeSpan := func() {
		if open {
			sb.WriteString("</span>")
			open = false
		}
	}
	openSpan := func() {
		if color == "" && !bold {
			return
		}
		var style []string
		if color != "" {
			style = append(style, "color:"+color)
		}
		if bold {
			style = append(style, "font-weight:bold")
		}
		sb.WriteString(`<span style="` + strings.Join(style, ";") + `">`)
		open = true
	}

	for len(s) > 0 {
		i := strings.Index(s, "\x1b[")
		if i < 0 {
			sb.WriteString(html.EscapeString(s))

			break
		}
		sb.WriteString(html.EscapeString(s[:i]))
		s = s[i+2:]

		end := strings.IndexByte(s, 'm')
		if end < 0 {
			break
		}
		codes := strings.Split(s[:end], ";")
		s = s[end+1:]

		closeSpan()
		for j := 0; j < len(codes); j++ {
			switch codes[j] {
			case "", "0":
				color, bold = "", false
			case "1":
				bold = true
			case "22":
				bold = false
			case "39":
				color = ""
			case "38":
				if j+4 < len(codes) && codes[j+1] == "2" {
					r, _ := strconv.Atoi(codes[j+2])
					g, _ := strconv.Atoi(codes[j+3])
					b, _ := strconv.Atoi(codes[j+4])
					color = fmt.Sprintf("#%02x%02x%02x", r, g, b)
					j += 4
				} else if j+2 < len(codes) && codes[j+1] == "5" {
					n, _ := strconv.Atoi(codes[j+2])
					color = ansi256[n]
					j += 2
				}
			}
		}
		openSpan()
	}
	closeSpan()

	return sb.String()
}
