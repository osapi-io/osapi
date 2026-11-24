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

type PrinterPublicTestSuite struct {
	suite.Suite
}

func (s *PrinterPublicTestSuite) TestPrinter() {
	p := cli.Printer{}

	tests := []struct {
		name string
		do   func()
		want []string
	}{
		{
			name: "when a block prints, it has its heading and its pairs",
			do: func() {
				p.Heading("Agent")
				p.KV("Hostname", "web-01")
				cli.FlushKV()
			},
			want: []string{"Agent:", "Hostname:", "web-01"},
		},
		{
			name: "when a table prints, it has its title, headers and rows",
			do: func() {
				p.Table(
					command.Table{
						Title:   "Files (1)",
						Headers: []string{"NAME"},
						Rows:    [][]string{{"motd"}},
					},
				)
			},
			want: []string{"Files (1)", "NAME", "motd"},
		},
		{
			name: "when a table is empty, it says None. and its hint",
			do: func() {
				p.Table(
					command.Table{
						Title:   "Files (0)",
						Headers: []string{"NAME"},
						Empty:   "Upload one.",
					},
				)
			},
			want: []string{"Files (0)", "None. Upload one."},
		},
		{
			name: "when a table is empty with no hint, it says None.",
			do: func() {
				p.Table(command.Table{Title: "Files (0)", Headers: []string{"NAME"}})
			},
			want: []string{"None."},
		},
		{
			name: "when a block prints, it is its title, then each pair under it",
			do:   func() { p.Block("Flat()", "", "Adds no word.") },
			want: []string{"Flat()", "Adds no word."},
		},
		{
			name: "when text prints, it is as it is, with one newline at its end",
			do:   func() { p.Text("# Title\n\ntext\n\n") },
			want: []string{"# Title\n\ntext\n"},
		},
		{
			name: "when a verdict prints, it is one line",
			do:   func() { p.Verdict("m-1 admitted") },
			want: []string{"m-1 admitted"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := cli.Capture(tc.do)
			for _, w := range tc.want {
				s.Contains(got, w)
			}
		})
	}

	s.Run("when text prints, no trace line follows it, so a pipe gets only the text", func() {
		cli.PrintTraceID("26e448c9dbe1c4a8d216a98c675079c9")
		got := cli.Capture(func() {
			p.Text("s3cret")
			cli.FlushTraceID()
		})
		s.Equal("s3cret\n", got)
	})

	s.Contains(p.Status("ok"), "ok")
	s.Equal(cli.FormatBytes(3480), p.Bytes(3480))
	s.Equal(cli.FormatAge(3*time.Second), p.Age(3*time.Second))
	at := time.Date(2026, 11, 12, 18, 15, 13, 0, time.UTC)
	s.Equal(cli.LocalTime(at), p.Time(at))
	s.Equal(cli.FormatDuration("42ms"), p.Duration(42*time.Millisecond))
}

func TestPrinterPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PrinterPublicTestSuite))
}
