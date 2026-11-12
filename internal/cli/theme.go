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
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The logo palette, from asset/logo-dark.svg. The banner is drawn in it and
// headings take their colors from it, so the CLI reads like the README.
var (
	LogoGray    = lipgloss.Color("#8a8a8a")
	LogoCyan    = lipgloss.Color("#00d5ff")
	LogoGreen   = lipgloss.Color("#45ae17")
	LogoYellow  = lipgloss.Color("#ffd100")
	LogoMagenta = lipgloss.Color("#ff00ff")
	LogoBlue    = lipgloss.Color("#2300fd")
	// LogoBlueFaint is the logo's blue at 25% opacity over a dark terminal.
	LogoBlueFaint = lipgloss.Color("#120c4c")
	// LogoBlueText is the logo's blue lifted to read as text on a dark
	// terminal. The logo's own #2300fd is nearly invisible on black.
	LogoBlueText = lipgloss.Color("#6b5cff")
)

// Status colors. They match the web UI, and they mean the same thing in every
// table, so they do not rotate with the headings.
var (
	Green  = lipgloss.Color("#67ea94") // UI --color-status-ready
	Yellow = lipgloss.Color("222")     // warm amber
	Red    = lipgloss.Color("#ef4444") // UI --color-status-error
)

// The styles every renderer in this package uses. Colors are defined in this
// file and nowhere else.
var (
	// HeadingStyle is the base for help section headings and output titles.
	// Each one takes its color from nextHeadingColor.
	HeadingStyle = lipgloss.NewStyle().Bold(true)
	// NameStyle is for what a reader scans for: command and flag names,
	// table column headers and key-value labels.
	NameStyle = lipgloss.NewStyle().Bold(true)
	// ValueStyle is for data. It is the terminal's own color.
	ValueStyle = lipgloss.NewStyle()
	// DimStyle is for secondary text.
	DimStyle = lipgloss.NewStyle().Foreground(LogoGray)
	// OKStyle, WarnStyle and ErrStyle carry a result's status.
	OKStyle   = lipgloss.NewStyle().Foreground(Green)
	WarnStyle = lipgloss.NewStyle().Foreground(Yellow)
	ErrStyle  = lipgloss.NewStyle().Foreground(Red)
)

// titleStyle is for a section title in command output.
func titleStyle() lipgloss.Style {
	return HeadingStyle.Foreground(nextHeadingColor())
}

// headingColors are the logo colors that read as text on a dark terminal.
// Blue is last so it is only reached by a fifth heading. Gray is secondary
// text, and red is never a heading because it means an error.
var headingColors = []lipgloss.Color{
	LogoCyan,
	LogoMagenta,
	LogoYellow,
	LogoGreen,
	LogoBlueText,
}

// headingOrder is headingColors shuffled once per run, so each run looks
// different but no two headings in one output share a color until all five
// are used.
var headingOrder = shuffled(headingColors)

// headingNext is the index of the next heading color in headingOrder.
var headingNext int

func shuffled(
	colors []lipgloss.Color,
) []lipgloss.Color {
	out := append([]lipgloss.Color(nil), colors...)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })

	return out
}

// nextHeadingColor returns the next color for a heading, wrapping after all
// five have been used.
func nextHeadingColor() lipgloss.Color {
	c := headingOrder[headingNext%len(headingOrder)]
	headingNext++

	return c
}

// SetHeadingOrder fixes the heading colors and restarts them, for tests.
func SetHeadingOrder(
	colors []lipgloss.Color,
) {
	headingOrder = colors
	headingNext = 0
}

// labelStyle is for table headers and key-value labels.
func labelStyle() lipgloss.Style { return NameStyle }

// rendererFor binds lipgloss to the writer being printed to, so NO_COLOR and
// a non-terminal destination give plain text for that writer rather than for
// stdout.
func rendererFor(
	w io.Writer,
) *lipgloss.Renderer {
	if f, ok := w.(*os.File); ok {
		return lipgloss.NewRenderer(f)
	}

	return lipgloss.DefaultRenderer()
}

func render(
	w io.Writer,
	st lipgloss.Style,
	s string,
) string {
	return st.Renderer(rendererFor(w)).Render(s)
}

// colorEnabled reports whether output to w is styled at all.
func colorEnabled(
	w io.Writer,
) bool {
	probe := render(w, lipgloss.NewStyle().Foreground(LogoCyan), "x")

	return strings.ContainsRune(probe, 0x1b)
}

// Mute renders secondary text in the logo's gray.
func Mute(
	w io.Writer,
	s string,
) string {
	return render(w, DimStyle, s)
}

// OK renders s in the success color.
func OK(
	w io.Writer,
	s string,
) string {
	return render(w, OKStyle, s)
}

// Err renders s in the error color.
func Err(
	w io.Writer,
	s string,
) string {
	return render(w, ErrStyle, s)
}

// Success renders msg after a green check, or "[ok]" without color.
func Success(
	w io.Writer,
	msg string,
) string {
	if !colorEnabled(w) {
		return "[ok] " + msg
	}

	return OK(w, "✓") + " " + msg
}

// Failure renders msg after a red cross, or "[err]" without color.
func Failure(
	w io.Writer,
	msg string,
) string {
	if !colorEnabled(w) {
		return "[err] " + msg
	}

	return Err(w, "✗") + " " + msg
}

// Heading renders a help section heading.
func Heading(
	w io.Writer,
	s string,
) string {
	return render(w, titleStyle(), s)
}

// Sub renders a command or flag name.
func Sub(
	w io.Writer,
	s string,
) string {
	return render(w, NameStyle, s)
}

// flagName matches the flag names at the start of a pflag usage line:
// "  -j, --json" or "      --osapi-file string".
var flagName = regexp.MustCompile(`(?m)^(\s+)((?:-\w, )?--[\w-]+)( [a-zA-Z]+)?`)

// Flags bolds the flag names in a pflag usage block and mutes their types.
func Flags(
	w io.Writer,
	usages string,
) string {
	if !colorEnabled(w) {
		return usages
	}

	return flagName.ReplaceAllStringFunc(usages, func(m string) string {
		parts := flagName.FindStringSubmatch(m)

		return parts[1] + Sub(w, parts[2]) + Mute(w, parts[3])
	})
}

// logo is the README logo as a grid of cells, row by row, from
// asset/logo-dark.svg: 17 columns of 16px cells, 4 rows. A dot is empty.
var logo = []string{
	"GGG.NG..GG.GGG.G.",
	"G.G.N..G.G.G.G.G.",
	"C.N..M.BBb.YYY.M.",
	"CCN.YM.B.b.Y...M.",
}

var logoColors = map[byte]lipgloss.Color{
	'G': LogoGray,
	'C': LogoCyan,
	'N': LogoGreen,
	'Y': LogoYellow,
	'M': LogoMagenta,
	'B': LogoBlue,
	'b': LogoBlueFaint,
}

// Banner draws the README logo one grid row per terminal line, each cell two
// columns of full block. A terminal cell is about twice as tall as it is
// wide, so two columns make a square, as in the SVG. Without color it returns
// nothing, because the logo is only shape and color.
func Banner(
	w io.Writer,
) string {
	if !colorEnabled(w) {
		return ""
	}

	var b strings.Builder
	for _, row := range logo {
		b.WriteString("  ")
		for col := 0; col < len(row); col++ {
			c, ok := logoColors[row[col]]
			if !ok {
				b.WriteString("  ")

				continue
			}
			b.WriteString(render(w, lipgloss.NewStyle().Foreground(c), "██"))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// PrintJSON prints a response body as it came from the API, for --json.
func PrintJSON(
	body []byte,
) {
	_, _ = fmt.Fprintln(os.Stdout, string(body))
}

// PrintNote prints a one-line message where a table would be, such as an
// empty result, indented like the tables and in the secondary color.
func PrintNote(
	msg string,
) {
	_, _ = fmt.Fprintln(os.Stdout, "  "+Mute(os.Stdout, msg))
}
