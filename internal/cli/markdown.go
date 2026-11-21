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

package cli

import (
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/guide"
)

// Markdown prints an output as a page of the docs site: a heading for each
// block, a table for each key-value block and each table, and a paragraph
// for each block of text. It writes what the terminal Printer writes, from
// the same calls, so a page and its command show the same thing.
type Markdown struct {
	// Title, when set, is the text of the first heading, in place of the
	// output's own, such as the name of a method for its section.
	Title string
	// Link, when set, is the page that a cell of a table links to, by its
	// column's header and its text, or "" for none: the name of each
	// behavior in a list links to its page.
	Link func(header string, cell string) string

	sb strings.Builder
	kv []string
}

var _ command.Printer = (*Markdown)(nil)

// HeadingLink is the header that Link gets for the text of a heading, so a
// heading such as the command of a section can link to its page.
const HeadingLink = "#"

// String is the page so far.
func (m *Markdown) String() string {
	m.flush()

	return m.sb.String()
}

// Heading starts a block.
func (m *Markdown) Heading(
	title string,
) {
	m.flush()
	if m.Title != "" {
		title, m.Title = m.Title, ""
	}
	title = strings.TrimSuffix(title, ":")
	text := guide.MDX(title)
	if m.Link != nil {
		if url := m.Link(HeadingLink, title); url != "" {
			text = "[" + text + "](" + url + ")"
		}
	}
	m.sb.WriteString("\n## " + text + "\n")
}

// KV adds pairs to the key-value block, which prints as a table.
func (m *Markdown) KV(
	pairs ...string,
) {
	m.kv = append(m.kv, pairs...)
}

// Table prints t as a table, under its title.
func (m *Markdown) Table(
	t command.Table,
) {
	m.flush()
	if t.Title != "" {
		m.sb.WriteString("\n### " + guide.MDX(strings.TrimSuffix(t.Title, ":")) + "\n")
	}
	if len(t.Rows) == 0 {
		m.sb.WriteString("\nNone. " + guide.MDX(t.Empty) + "\n")

		return
	}
	m.sb.WriteString("\n|")
	for _, h := range t.Headers {
		m.sb.WriteString(" " + cellMD(h) + " |")
	}
	m.sb.WriteString("\n|")
	for range t.Headers {
		m.sb.WriteString(" --- |")
	}
	m.sb.WriteString("\n")
	for _, row := range t.Rows {
		m.sb.WriteString("|")
		for i, c := range row {
			header := ""
			if i < len(t.Headers) {
				header = t.Headers[i]
			}
			m.sb.WriteString(" " + m.linked(header, c) + " |")
		}
		m.sb.WriteString("\n")
	}
}

// Verdict prints text as a paragraph.
func (m *Markdown) Verdict(
	text string,
) {
	m.flush()
	m.sb.WriteString("\n" + guide.MDX(text) + "\n")
}

// Block prints one item of a list of blocks: its title as a heading, then
// each pair as a paragraph, "Label: text", or the text alone.
func (m *Markdown) Block(
	title string,
	pairs ...string,
) {
	m.flush()
	m.sb.WriteString("\n### " + guide.MDX(title) + "\n")
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == "" {
			m.sb.WriteString("\n" + guide.MDX(pairs[i+1]) + "\n")

			continue
		}
		m.sb.WriteString("\n" + guide.MDX(pairs[i]) + ": " + guide.MDX(pairs[i+1]) + "\n")
	}
}

// Text prints text as a paragraph.
func (m *Markdown) Text(
	text string,
) {
	m.flush()
	m.sb.WriteString("\n" + guide.MDX(strings.TrimRight(text, "\n")) + "\n")
}

// Status is word, with no color.
func (m *Markdown) Status(
	word string,
) string {
	return word
}

// Bytes is n as a size, as the terminal shows it.
func (m *Markdown) Bytes(
	n int64,
) string {
	return FormatBytes(int(n))
}

// Age is d as an age, as the terminal shows it.
func (m *Markdown) Age(
	d time.Duration,
) string {
	return FormatAge(d)
}

// Time is t in UTC, so that a page is the same wherever it is written.
func (m *Markdown) Time(
	t time.Time,
) string {
	return t.UTC().Format(time.RFC3339)
}

// Duration is d rounded, as the terminal shows it.
func (m *Markdown) Duration(
	d time.Duration,
) string {
	return FormatDuration(d.String())
}

// flush prints the key-value block, if there is one, as a table of two
// columns.
func (m *Markdown) flush() {
	if len(m.kv) == 0 {
		return
	}
	m.sb.WriteString("\n| | |\n| --- | --- |\n")
	for i := 0; i+1 < len(m.kv); i += 2 {
		m.sb.WriteString("| " + cellMD(m.kv[i]) + " | " + m.linked(m.kv[i], m.kv[i+1]) + " |\n")
	}
	m.kv = nil
}

// cellMD is s safe in a cell of a table on an MDX page: on one line, with a
// pipe escaped and markup escaped outside code.
func cellMD(
	s string,
) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "|", `\|`)

	return guide.MDX(s)
}

// linked is the text of a cell, with each value that Link knows as a link:
// a list such as netplan, managedfile links each of its values.
func (m *Markdown) linked(
	header string,
	text string,
) string {
	if m.Link == nil {
		return cellMD(text)
	}
	values := strings.Split(text, ", ")
	for i, v := range values {
		if url := m.Link(header, v); url != "" {
			values[i] = "[" + cellMD(v) + "](" + url + ")"
		} else {
			values[i] = cellMD(v)
		}
	}

	return strings.Join(values, ", ")
}
