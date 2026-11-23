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

package cli_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"
)

type MarkdownPublicTestSuite struct {
	suite.Suite
}

func (s *MarkdownPublicTestSuite) TestPage() {
	m := &cli.Markdown{
		Title: "Method get",
		Link: func(header string, cell string) string {
			switch {
			case header == cli.HeadingLink && cell == "Linked":
				return "linked.md"
			case (header == "NAME" || header == "Uses") && cell != "none":
				return cell + ".md"
			}

			return ""
		},
	}
	m.Heading("Method:")
	m.KV("Name", "get", "Uses", "netplan, managedfile", "Note", "a | b\nc",
		"Writes", "/etc/cron.d/*, /etc/cron.daily/*")
	m.Heading("Linked")
	m.Table(command.Table{
		Title: "Methods (1):", Headers: []string{"NAME", "ROUTE"},
		Rows: [][]string{{"dns", "/api/{hostname}"}, {"none"}},
	})
	m.Table(command.Table{Title: "Settings (0):", Empty: "Declare them."})
	m.Table(command.Table{Headers: []string{"X"}, Rows: [][]string{{"y"}}})
	m.Verdict("done <now>")
	m.Block("Term", "", "Alone.", "Use", "On {get}.")
	m.Text("plain text\n")

	page := m.String()
	s.Contains(page, "\n## Method get\n", "the first heading is the title")
	s.Contains(page, "| Uses | [netplan](netplan.md), [managedfile](managedfile.md) |",
		"each value of a list links")
	s.Contains(page, "| Note | a \\| b c |", "a pipe is escaped and a newline is a space")
	s.Contains(page, `| Writes | /etc/cron.d/\*, /etc/cron.daily/\* |`,
		"an asterisk is escaped, so two are not emphasis")
	s.Contains(page, "\n## [Linked](linked.md)\n", "a heading links")
	s.Contains(page, "\n### Methods (1)\n\n| NAME | ROUTE |\n| --- | --- |\n")
	s.Contains(page, "| [dns](dns.md) | /api/&#123;hostname&#125; |")
	s.Contains(page, "| none |", "a value Link does not know stays text, and a short row its cells")
	s.Contains(page, "\n### Settings (0)\n\nNone. Declare them.\n")
	s.Contains(page, "\n| X |\n| --- |\n| y |\n", "a table with no title")
	s.Contains(page, "\ndone &lt;now&gt;\n")
	s.Contains(page, "\n### Term\n\nAlone.\n\nUse: On &#123;get&#125;.\n")
	s.Contains(page, "\nplain text\n")
}

func (s *MarkdownPublicTestSuite) TestFormats() {
	m := &cli.Markdown{}
	s.Equal("ok", m.Status("ok"), "a status has no color")
	s.Equal(cli.FormatBytes(2048), m.Bytes(2048))
	s.Equal(cli.FormatAge(time.Minute), m.Age(time.Minute))
	s.Equal(
		"2026-11-21T08:00:00Z",
		m.Time(time.Date(2026, 11, 21, 0, 0, 0, 0, time.FixedZone("PST", -8*3600))),
		"a time is in UTC, so a page is the same wherever it is written",
	)
	s.Equal(cli.FormatDuration((90 * time.Second).String()), m.Duration(90*time.Second))
	m.KV("A", "b")
	s.Contains(m.String(), "| A | b |", "a key-value block with no Link")
}

func TestMarkdownPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MarkdownPublicTestSuite))
}
