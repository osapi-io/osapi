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
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

// statusStyles maps a status word to the style that carries it. The word is
// the first one in the cell, lowercased, so "ok nats://localhost:4222" and
// "Ready" both match.
var statusStyles = map[string]lipgloss.Style{
	"ok":        OKStyle,
	"ready":     OKStyle,
	"changed":   OKStyle,
	"completed": OKStyle,
	"running":   OKStyle,
	"healthy":   OKStyle,
	"active":    OKStyle,
	"enabled":   OKStyle,
	"accepted":  OKStyle,
	"skip":      WarnStyle,
	"skipped":   WarnStyle,
	"disabled":  WarnStyle,
	"draining":  WarnStyle,
	"drained":   WarnStyle,
	"cordoned":  WarnStyle,
	"pending":   WarnStyle,
	"submitted": WarnStyle,
	"partial":   WarnStyle,
	"degraded":  WarnStyle,
	"unknown":   WarnStyle,
	"err":       ErrStyle,
	"error":     ErrStyle,
	"failed":    ErrStyle,
	"timeout":   ErrStyle,
	"notready":  ErrStyle,
	"unhealthy": ErrStyle,
	"rejected":  ErrStyle,
}

// statusStyle returns the style for a status cell, or ValueStyle when the
// cell does not start with a status word.
func statusStyle(
	cell string,
) lipgloss.Style {
	word, _, _ := strings.Cut(strings.TrimSpace(cell), " ")
	if s, ok := statusStyles[strings.ToLower(word)]; ok {
		return s
	}

	return ValueStyle
}

// countStyles maps what a count counts to the style it takes when above
// zero. A zero is dim whatever it counts: nothing happened.
var countStyles = map[string]lipgloss.Style{
	"completed":   OKStyle,
	"ok":          OKStyle,
	"changed":     OKStyle,
	"failed":      ErrStyle,
	"dlq":         ErrStyle,
	"errors":      ErrStyle,
	"submitted":   WarnStyle,
	"unprocessed": WarnStyle,
	"pending":     WarnStyle,
	"partial":     WarnStyle,
	"skipped":     WarnStyle,
}

// Count renders a count of kind, such as 3 failed, in the style of what it
// counts: green for completed, red for failed, yellow for waiting, plain for
// a total, and dim for zero.
func Count(
	n int,
	kind string,
) string {
	return CountValue(n, kind) + " " + kind
}

// CountValue renders just the number of a count, for a key-value pair whose
// label already says what it counts.
func CountValue(
	n int,
	kind string,
) string {
	s := fmt.Sprintf("%d", n)
	if n == 0 {
		return DimStyle.Render(s)
	}
	if style, ok := countStyles[strings.ToLower(kind)]; ok {
		return style.Render(s)
	}

	return ValueStyle.Render(s)
}

// treeBranches are the prefixes a nested row's first cells start with.
var treeBranches = []string{"├─ ", "└─ ", "│  "}

// renderCell styles one table cell by its column, so every table colors the
// same columns the same way:
//
//   - STATUS takes its status's color.
//   - HOSTNAME is bold, what a reader scans for, like a command name in help.
//     A nested row's tree branch is dimmed and its name left plain.
//   - CONDITIONS is yellow when a condition is set and dim when it is "-".
//   - An ID column is dim: it is for copying, not reading. AGE and CREATED
//     are dim too: when, not what.
//   - TYPE and NAME are bold, what a reader scans for.
//   - METHOD is an HTTP verb, bold. ROUTE and PATH bold their verb and gray
//     their {params}. VERB is bold for a write, which changes the host.
//   - PERMISSION, TIMESTAMP, UPDATED and SINCE are gray, like IDs.
//   - Everything else is in the value style.
func renderCell(
	header string,
	cell string,
) string {
	h := strings.ToUpper(header)

	switch {
	case h == "STATUS":
		return statusStyle(cell).Render(cell)
	case h == "HOSTNAME":
		for _, b := range treeBranches {
			if rest, ok := strings.CutPrefix(cell, b); ok {
				return DimStyle.Render(b) + ValueStyle.Render(rest)
			}
		}

		return NameStyle.Render(cell)
	case h == "CONDITIONS":
		if strings.TrimSpace(cell) == "" || cell == "-" {
			return DimStyle.Render(cell)
		}

		return WarnStyle.Render(cell)
	case h == "ID" || strings.HasSuffix(h, " ID"), h == "AGE", h == "CREATED",
		h == "TIMESTAMP", h == "UPDATED", h == "SINCE", h == "PERMISSION":
		return DimStyle.Render(cell)
	case h == "TYPE", h == "NAME":
		return NameStyle.Render(cell)
	case h == "METHOD":
		return NameStyle.Render(cell)
	case h == "VERB":
		if strings.EqualFold(strings.TrimSpace(cell), "write") {
			return NameStyle.Render(cell)
		}

		return ValueStyle.Render(cell)
	case h == "ROUTE" || h == "PATH":
		return renderRoute(cell)
	default:
		return ValueStyle.Render(cell)
	}
}

// tints are the dark tints of the heading colors. A block's table headers
// and key-value labels take the tint of its heading's color, so a block reads
// as one unit and labels sit a step below the heading that owns them.
var tints = map[lipgloss.Color]lipgloss.Color{
	LogoCyan:     lipgloss.Color("#0a8fb0"),
	LogoMagenta:  lipgloss.Color("#b030b0"),
	LogoYellow:   lipgloss.Color("#b08f00"),
	LogoGreen:    lipgloss.Color("#3a8f14"),
	LogoBlueText: lipgloss.Color("#5a4cd8"),
}

// currentTint is the tint of the most recent heading's color, or empty
// before any heading.
var currentTint lipgloss.Color

// titleStyle is for a section title in command output. It sets the tint the
// block's labels take.
func titleStyle() lipgloss.Style {
	return headingStyleFor(nextHeadingColor())
}

func headingStyleFor(
	c lipgloss.Color,
) lipgloss.Style {
	currentTint = tints[c]

	return HeadingStyle.Foreground(c)
}

// titleStyleAbove is titleStyle for a title over text in the given status
// styles ("ok", "warn"), never in a color that text is already in.
func titleStyleAbove(
	statuses ...string,
) lipgloss.Style {
	avoid := make([]lipgloss.Color, 0, len(statuses))
	for _, s := range statuses {
		if c, ok := clashes[s]; ok {
			avoid = append(avoid, c)
		}
	}

	return headingStyleFor(nextHeadingColor(avoid...))
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
// five have been used. A heading is never the color of the text under it, so
// a color in avoid is skipped: the next one that is not is swapped into its
// place, which keeps two headings in one output from sharing a color.
func nextHeadingColor(
	avoid ...lipgloss.Color,
) lipgloss.Color {
	n := len(headingOrder)
	at := headingNext % n
	for k := range n {
		i := (at + k) % n
		if !slices.Contains(avoid, headingOrder[i]) {
			headingOrder[at], headingOrder[i] = headingOrder[i], headingOrder[at]

			break
		}
	}
	headingNext++

	return headingOrder[at]
}

// clashes is the logo color each status style would be mistaken for in a
// heading: yellow above a skip reads as one block, as does green above ok.
var clashes = map[string]lipgloss.Color{
	"warn": LogoYellow,
	"ok":   LogoGreen,
}

// pipe makes the pipe Capture reads through; a test swaps it to fail.
var pipe = os.Pipe

// Capture runs fn with stdout redirected and color forced on, and returns what
// it printed with the theme's escape codes intact. The heading colors start
// from the palette's own order rather than the run's shuffle, so the same
// output captures the same way every time. The docs generator uses it to put
// the terminal's exact output on a page.
func Capture(
	fn func(),
) string {
	prevOrder, prevNext := headingOrder, headingNext
	SetHeadingOrder(append([]lipgloss.Color(nil), headingColors...))
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	defer func() {
		lipgloss.SetColorProfile(termenv.ColorProfile())
		headingOrder, headingNext = prevOrder, prevNext
	}()

	r, w, err := pipe()
	if err != nil {
		return ""
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	FlushKV()

	_ = w.Close()
	os.Stdout = orig

	return <-done
}

// SetHeadingOrder fixes the heading colors and restarts them, for tests.
func SetHeadingOrder(
	colors []lipgloss.Color,
) {
	headingOrder = colors
	headingNext = 0
}

// labelStyle is for table headers and key-value labels.
func labelStyle() lipgloss.Style {
	if currentTint == "" {
		return DimStyle
	}

	return lipgloss.NewStyle().Foreground(currentTint)
}

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
	FlushKV()
	_, _ = fmt.Fprintln(os.Stdout, string(body))
}

// PrintNote prints a one-line message where a table would be, such as an
// empty result, indented like the tables and in the secondary color.
func PrintNote(
	msg string,
) {
	FlushKV()
	_, _ = fmt.Fprintln(os.Stdout, "  "+Mute(os.Stdout, msg))
}

// httpVerbs are the methods a ROUTE cell can start with.
var httpVerbs = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

// renderRoute bolds a route's HTTP verb and grays its {params}, so the shape
// of the path reads before its variables.
func renderRoute(
	cell string,
) string {
	var sb strings.Builder
	rest := cell
	if verb, path, ok := strings.Cut(cell, " "); ok && httpVerbs[verb] {
		sb.WriteString(NameStyle.Render(verb) + " ")
		rest = path
	}
	for len(rest) > 0 {
		i := strings.IndexByte(rest, '{')
		j := strings.IndexByte(rest, '}')
		if i < 0 || j < i {
			sb.WriteString(ValueStyle.Render(rest))

			break
		}
		sb.WriteString(ValueStyle.Render(rest[:i]))
		sb.WriteString(DimStyle.Render(rest[i : j+1]))
		rest = rest[j+1:]
	}

	return sb.String()
}
