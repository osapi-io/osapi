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
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Section represents a header with its corresponding rows.
type Section struct {
	Title    string
	Headers  []string
	Rows     [][]string
	Errors   []ErrorEntry
	Duration string // e.g. "286ms" — shown in summary line
	// Reference marks a table that is not per-host results, such as a
	// behavior's methods, so no host summary line is printed under it.
	Reference bool
	// Empty is what an empty list says instead of rows: what to do next,
	// such as "None. Upload one with osapi client file upload." Defaults to
	// "None.".
	Empty string
}

// ResultRow is a per-host result used by BuildBroadcastTable and
// BuildMutationTable.
type ResultRow struct {
	Hostname string
	Status   string
	Changed  *bool
	Error    *string
	Fields   []string
}

// ErrorEntry is an error or skip reason from a host, rendered below the table.
type ErrorEntry struct {
	Hostname string
	Message  string
	Status   string // "err" or "skip"
}

// TableResult holds the output of BuildBroadcastTable / BuildMutationTable.
type TableResult struct {
	Headers []string
	Rows    [][]string
	Errors  []ErrorEntry
}

// resolveStatus computes the compact STATUS value from a ResultRow.
// Values: ok, changed, skip, err.
func resolveStatus(
	r ResultRow,
) string {
	// Check skipped before error — skipped operations set an error
	// message ("unsupported on this OS family") but are not failures.
	if r.Status == "skipped" || r.Status == "skip" {
		return "skip"
	}

	if r.Error != nil {
		return "err"
	}

	if r.Changed != nil && *r.Changed {
		return "changed"
	}

	return "ok"
}

// BuildBroadcastTable builds a TableResult for a broadcast response.
// HOSTNAME and STATUS are always shown. Errors are collected for
// rendering below the table by PrintCompactTable.
func BuildBroadcastTable(
	results []ResultRow,
	fieldHeaders []string,
) TableResult {
	return buildTable(results, fieldHeaders)
}

// BuildMutationTable builds a TableResult for a mutation response.
// Uses the same unified STATUS column (ok/changed/skip/err).
func BuildMutationTable(
	results []ResultRow,
	fieldHeaders []string,
) TableResult {
	return buildTable(results, fieldHeaders)
}

// buildTable is the shared implementation for broadcast and mutation tables.
func buildTable(
	results []ResultRow,
	fieldHeaders []string,
) TableResult {
	headers := make([]string, 0, 2+len(fieldHeaders))
	headers = append(headers, "HOSTNAME", "STATUS")
	headers = append(headers, fieldHeaders...)

	var errors []ErrorEntry
	rows := make([][]string, 0, len(results))

	// Hosts print in name order, so a fleet prints the same way every run.
	sorted := append([]ResultRow(nil), results...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Hostname < sorted[j].Hostname })

	for _, r := range sorted {
		status := resolveStatus(r)

		// A host that was skipped or failed has no data to show, so it is
		// listed once under Details with its reason rather than as a row of
		// empty columns.
		if r.Error != nil || status == "skip" || status == "err" {
			msg := ""
			if r.Error != nil {
				msg = *r.Error
			}
			errors = append(errors, ErrorEntry{
				Hostname: r.Hostname,
				Message:  msg,
				Status:   status,
			})

			continue
		}

		row := []string{r.Hostname, status}
		row = append(row, r.Fields...)
		rows = append(rows, row)
	}

	return TableResult{
		Headers: headers,
		Rows:    rows,
		Errors:  errors,
	}
}

// BoolToSafeString converts a *bool to a string. Returns "" if nil.
func BoolToSafeString(
	b *bool,
) string {
	if b != nil {
		return fmt.Sprintf("%v", *b)
	}
	return ""
}

// PrintHeading prints a section heading the way help prints one: a blank
// line, then the title and a colon at the left margin in the next logo color,
// with the block it heads indented under it.
func PrintHeading(
	title string,
) {
	printHeading(titleStyle(), title)
}

// headed is set by a heading and cleared by the block under it, so an
// untitled table straight after a heading does not open with a blank line.
var headed bool

func printHeading(
	style lipgloss.Style,
	title string,
) {
	FlushKV()
	fmt.Printf("\n%s\n", style.Render(title+":"))
	headed = true
}

// compactMaxColWidth is the maximum column width before truncation.
const compactMaxColWidth = 79

// PrintCompactTable renders a compact column-aligned table (kubectl-style).
// Headers are uppercase and bold, data rows are in the value style, with a
// 2-space indent.
// Multi-line cell values are flattened to a single line and long values
// are truncated with an ellipsis.
func PrintCompactTable(
	sections []Section,
) {
	FlushKV()
	const colGap = 2

	for _, section := range sections {
		// A table of per-host results is headed Results unless it says
		// otherwise, so every command's output reads like help: a colored
		// heading, then the block indented under it.
		title := section.Title
		if title == "" && len(section.Headers) > 0 &&
			strings.EqualFold(section.Headers[0], "hostname") {
			title = "Results"
		}
		switch {
		case title != "":
			PrintHeading(title)
		case !headed:
			fmt.Println()
		}
		headed = false

		// An empty table prints no header row. Per-host results go
		// straight to their summary and details; any other list says so.
		if len(section.Rows) == 0 {
			if len(section.Errors) == 0 {
				empty := section.Empty
				if empty == "" {
					empty = "None."
				}
				fmt.Println("  " + DimStyle.Render(empty))
			}
			printSummary(section)
			PrintErrors(section.Errors)

			continue
		}

		headerStyle := labelStyle()

		// Flatten multi-line cells to single lines for compact display
		flatRows := make([][]string, len(section.Rows))
		for r, row := range section.Rows {
			flat := make([]string, len(row))
			for c, cell := range row {
				flat[c] = strings.Join(strings.Fields(cell), " ")
			}
			flatRows[r] = flat
		}

		// Calculate column widths from headers and flattened data,
		// capping at compactMaxColWidth to prevent blown-out columns.
		widths := make([]int, len(section.Headers))
		for i, h := range section.Headers {
			widths[i] = len(h)
		}
		for _, row := range flatRows {
			for i, cell := range row {
				if i < len(widths) && len(cell) > widths[i] {
					widths[i] = len(cell)
				}
			}
		}
		for i := range widths {
			if widths[i] > compactMaxColWidth {
				widths[i] = compactMaxColWidth
			}
		}

		// Build header line
		var hdr strings.Builder
		hdr.WriteString("  ")
		for i, h := range section.Headers {
			if i < len(section.Headers)-1 {
				hdr.WriteString(
					headerStyle.Render(fmt.Sprintf("%-*s", widths[i]+colGap, strings.ToUpper(h))),
				)
			} else {
				hdr.WriteString(headerStyle.Render(strings.ToUpper(h)))
			}
		}
		fmt.Println(hdr.String())

		// Build data rows.
		for r, row := range flatRows {
			var line strings.Builder
			line.WriteString("  ")
			for i := range section.Headers {
				cell := ""
				if i < len(row) {
					cell = row[i]
				}
				// Truncate cells that exceed the column width
				if len(cell) > widths[i] {
					cell = cell[:widths[i]-1] + "…"
				}
				// Style the cell, then pad it, so the padding is never colored
				// and a cell styled in two parts still lines up. A hostname
				// repeated from the row above is gray, so each host's rows
				// read as one block.
				if i == 0 && r > 0 && strings.EqualFold(section.Headers[0], "hostname") &&
					len(flatRows[r-1]) > 0 && flatRows[r-1][0] == cell {
					line.WriteString(DimStyle.Render(cell))
				} else {
					line.WriteString(renderCell(section.Headers[i], cell))
				}
				if i < len(section.Headers)-1 {
					line.WriteString(
						strings.Repeat(
							" ",
							max(widths[i]+colGap-utf8.RuneCountInString(cell), colGap),
						),
					)
				}
			}
			fmt.Println(line.String())
		}

		printSummary(section)
		PrintErrors(section.Errors)
	}
}

// printSummary renders a status summary line below the table.
// Format: "2 hosts: 1 ok, 1 skipped in 286ms"
func printSummary(
	section Section,
) {
	if section.Reference {
		return
	}

	// Count unique hostnames and their statuses from the rows.
	// STATUS is always the second column (index 1).
	hostStatuses := make(map[string]string)
	for _, row := range section.Rows {
		if len(row) >= 2 {
			hostname := row[0]
			status := row[1]
			// Keep the "worst" status per host.
			if cur, ok := hostStatuses[hostname]; !ok || statusWeight(status) > statusWeight(cur) {
				hostStatuses[hostname] = status
			}
		}
	}

	// Count hosts listed under Details by their own status.
	for _, e := range section.Errors {
		if _, ok := hostStatuses[e.Hostname]; !ok {
			status := e.Status
			if status != "skip" {
				status = "err"
			}
			hostStatuses[e.Hostname] = status
		}
	}

	totalHosts := len(hostStatuses)
	if totalHosts == 0 {
		return
	}

	counts := map[string]int{}
	for _, status := range hostStatuses {
		counts[status]++
	}

	var parts []string
	if n := counts["ok"]; n > 0 {
		parts = append(parts, OKStyle.Render(fmt.Sprintf("%d ok", n)))
	}
	if n := counts["changed"]; n > 0 {
		parts = append(parts, OKStyle.Render(fmt.Sprintf("%d changed", n)))
	}
	if n := counts["skip"]; n > 0 {
		parts = append(parts, WarnStyle.Render(fmt.Sprintf("%d skipped", n)))
	}
	if n := counts["err"]; n > 0 {
		parts = append(parts, ErrStyle.Render(fmt.Sprintf("%d failed", n)))
	}

	// A table whose rows are not per-host results, such as the agent list,
	// has nothing to count.
	if len(parts) == 0 {
		return
	}

	noun := "hosts"
	if totalHosts == 1 {
		noun = "host"
	}
	summary := fmt.Sprintf("%d %s: %s", totalHosts, noun, strings.Join(parts, ", "))
	if section.Duration != "" {
		summary += DimStyle.Render(fmt.Sprintf(" in %s", section.Duration))
	}
	if pendingJobID != "" {
		summary += "   " + DimStyle.Render("job "+pendingJobID)
		pendingJobID = ""
	}

	fmt.Printf("\n  %s\n", summary)
}

// pendingJobID is the job a result came from, set by PrintJobID and printed
// at the end of the result's summary line.
var pendingJobID string

// PrintJobID records the job a command's result came from. The result's
// summary line ends with it, in gray, rather than it floating above the
// Results heading.
func PrintJobID(
	id string,
) {
	pendingJobID = id
}

// statusWeight returns a numeric weight for status ordering.
// Higher = worse, used to pick the "worst" status per host.
func statusWeight(
	status string,
) int {
	switch status {
	case "ok":
		return 0
	case "changed":
		return 1
	case "skip":
		return 2
	case "err":
		return 3
	default:
		return 0
	}
}

// PrintErrors renders error and skip entries below a table. Errors are
// red, skips are yellow. Each entry shows hostname and message.
func PrintErrors(
	errors []ErrorEntry,
) {
	FlushKV()
	if len(errors) == 0 {
		return
	}

	// Find longest hostname for alignment.
	maxHost := 0
	for _, e := range errors {
		if len(e.Hostname) > maxHost {
			maxHost = len(e.Hostname)
		}
	}

	// Each entry says what happened, skipped or failed, in its status color,
	// then why. The title takes neither of the logo colors close to amber and
	// green, so it never matches the entries under it.
	printHeading(titleStyleAbove("warn", "ok"), "Details")
	for _, e := range errors {
		outcome, style := "failed", ErrStyle
		if e.Status == "skip" {
			outcome, style = "skipped", WarnStyle
		}
		fmt.Printf(
			"  %s  %s  %s\n",
			NameStyle.Render(fmt.Sprintf("%-*s", maxHost, e.Hostname)),
			style.Render(fmt.Sprintf("%-7s", outcome)),
			ValueStyle.Render(e.Message),
		)
	}
}

// PrintKV prints labeled key-value pairs on a single indented line.
// Pairs are padded to equal column widths for alignment.
// kvBlock holds key-value pairs until something else prints, so a block
// built from several PrintKV calls is aligned as one: every value starts in
// the same column.
var kvBlock [][2]string

// PrintKV adds key-value pairs to the current block. The block prints, one
// pair per line, a label in the block's tint, a colon, then the value lined up
// with the others, as soon as anything else prints or the command ends. A
// Status value takes its status's color.
func PrintKV(
	pairs ...string,
) {
	if len(pairs)%2 != 0 || len(pairs) == 0 {
		return
	}

	for i := 0; i < len(pairs); i += 2 {
		kvBlock = append(kvBlock, [2]string{pairs[i], pairs[i+1]})
	}
}

// FlushKV prints the pending key-value block. Every other printer calls it
// first, and cmd's Execute calls it when a command returns, so a block always
// prints before whatever follows it.
func FlushKV() {
	if len(kvBlock) == 0 {
		return
	}

	width := 0
	for _, kv := range kvBlock {
		if w := lipgloss.Width(kv[0]) + 1; w > width {
			width = w
		}
	}

	block := kvBlock
	kvBlock = nil
	for _, kv := range block {
		label, raw := kv[0], kv[1]
		value := ValueStyle.Render(raw)
		if strings.EqualFold(label, "status") {
			value = statusStyle(raw).Render(raw)
		}
		pad := strings.Repeat(" ", width-lipgloss.Width(label+":")+2)
		fmt.Println("  " + labelStyle().Render(label+":") + pad + value)
	}
	headed = false
}

// FormatAge formats a duration as a human-readable age string.
// Returns "3d 4h", "12h 30m", "45m", "30s" etc.
func FormatAge(
	d time.Duration,
) string {
	if d <= 0 {
		return ""
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
}

// FormatBytes formats a byte count as a human-readable string (e.g., "5.2 KB", "1.0 MB").
func FormatBytes(
	b int,
) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// FormatList helper function to convert []string to a formatted string.
func FormatList(
	list []string,
) string {
	if len(list) == 0 {
		return "None"
	}
	return strings.Join(list, ", ")
}

// FormatLabels formats a label map as "key:value, key:value" sorted by key.
func FormatLabels(
	labels map[string]string,
) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+":"+labels[k])
	}
	return strings.Join(parts, ", ")
}

// CalculateColumnWidths calculates the optimal width for each column based on content.
func CalculateColumnWidths(
	headers []string,
	rows [][]string,
	minPadding int,
) []int {
	if len(headers) == 0 {
		return []int{}
	}

	widths := make([]int, len(headers))

	// Start with header lengths
	for i, header := range headers {
		widths[i] = len(header)
	}

	// Check all row data to find max width per column
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				// For multi-line content, use the width of the longest line
				maxLineWidth := GetMaxLineWidth(cell)
				if maxLineWidth > widths[i] {
					widths[i] = maxLineWidth
				}
			}
		}
	}

	// Add padding to each column
	for i := range widths {
		widths[i] += minPadding * 2 // padding on both sides
	}

	return widths
}

// GetMaxLineWidth returns the width of the longest line in a multi-line string.
func GetMaxLineWidth(
	text string,
) int {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

// SafeString function to safely dereference string pointers.
func SafeString(
	s *string,
) string {
	if s != nil {
		return *s
	}
	return ""
}

// SafeUUID converts a *uuid.UUID to its string representation. Returns "" if nil.
func SafeUUID(
	u *uuid.UUID,
) string {
	if u != nil {
		return u.String()
	}
	return ""
}

// Float64ToSafeString converts a *float64 to a string. Returns "N/A" if nil.
func Float64ToSafeString(
	f *float64,
) string {
	if f != nil {
		return fmt.Sprintf("%f", *f)
	}
	return "N/A"
}

// IntToSafeString converts a *int to a string. Returns "N/A" if nil.
func IntToSafeString(
	i *int,
) string {
	if i != nil {
		return fmt.Sprintf("%d", *i)
	}
	return "N/A"
}

// HandleError reports err and exits 1: one line for a person, or under --json
// the error document with the API's status code named.
func HandleError(
	err error,
	_ *slog.Logger,
) {
	FlushKV()
	if !jsonOutput {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			msg := fmt.Sprintf("api error (%d): %s", apiErr.StatusCode, apiErr.Message)
			_, _ = fmt.Fprintln(errOut, Failure(errOut, msg))
		} else {
			_, _ = fmt.Fprintln(errOut, Failure(errOut, err.Error()))
		}
		osExit(1)

		return
	}

	doc := ErrorDocument{Error: err.Error(), Code: CodeError}
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		doc = ErrorDocument{
			Error:  apiErr.Message,
			Code:   APICode(apiErr.StatusCode),
			Status: apiErr.StatusCode,
		}
	}
	WriteErrorJSON(errOut, doc)
	osExit(ExitError)
}

// DisplayJobDetail displays detailed job information from domain types.
// Used by both job get and job run commands.
func DisplayJobDetail(
	resp *client.JobDetail,
) {
	// Display job metadata
	PrintHeading("Job")
	PrintKV("Job ID", resp.ID, "Status", resp.Status)
	if resp.Hostname != "" {
		PrintKV("Hostname", resp.Hostname)
	}
	if resp.Created != "" {
		PrintKV("Created", FormatTime(resp.Created))
	}
	if resp.UpdatedAt != "" {
		PrintKV("Updated", FormatTime(resp.UpdatedAt))
	}
	if resp.Error != "" {
		PrintKV("Error", resp.Error)
	}

	// Add agent summary from agent_states
	if len(resp.AgentStates) > 0 {
		completed := 0
		failed := 0
		skipped := 0
		processing := 0

		for _, state := range resp.AgentStates {
			switch state.Status {
			case string(job.StatusCompleted):
				completed++
			case string(job.StatusFailed):
				failed++
			case string(job.StatusSkipped):
				skipped++
			case string(job.StatusStarted):
				processing++
			}
		}

		total := len(resp.AgentStates)
		if total > 1 {
			summary := fmt.Sprintf(
				"%d total (%d completed, %d failed, %d processing",
				total,
				completed,
				failed,
				processing,
			)
			if skipped > 0 {
				summary += fmt.Sprintf(", %d skipped", skipped)
			}
			summary += ")"
			PrintKV("Agents", summary)
		}
	}

	// The request, as key-value pairs rather than raw JSON.
	if resp.Operation != nil {
		printObject("Request", resp.Operation)
	}

	var sections []Section

	// Display agent responses (for broadcast jobs)
	if len(resp.Responses) > 0 {
		responseRows := make([][]string, 0, len(resp.Responses))
		for _, hostname := range sortedKeys(resp.Responses) {
			response := resp.Responses[hostname]
			row := []string{hostname, response.Status, Summarize(response.Data), response.Error}
			responseRows = append(responseRows, row)
		}

		sections = append(sections, Section{
			Title:   "Agent Responses",
			Headers: []string{"HOSTNAME", "STATUS", "DATA", "ERROR"},
			Rows:    responseRows,
		})
	}

	// Display agent states (for broadcast jobs)
	if len(resp.AgentStates) > 0 {
		stateRows := make([][]string, 0, len(resp.AgentStates))
		for _, hostname := range sortedKeys(resp.AgentStates) {
			state := resp.AgentStates[hostname]
			stateRows = append(
				stateRows,
				[]string{hostname, state.Status, FormatDuration(state.Duration), state.Error},
			)
		}

		sections = append(sections, Section{
			Title:   "Agent States",
			Headers: []string{"HOSTNAME", "STATUS", "DURATION", "ERROR"},
			Rows:    stateRows,
		})
	}

	// Display timeline
	timelineRows := make([][]string, 0, len(resp.Timeline))
	for _, te := range resp.Timeline {
		timelineRows = append(
			timelineRows,
			[]string{te.Timestamp, te.Event, te.Hostname, te.Message, te.Error},
		)
	}
	sections = append(sections, Section{
		Title:   "Timeline",
		Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
		Rows:    timelineRows,
	})

	for _, sec := range sections {
		PrintCompactTable([]Section{sec})
	}

	// The result, as key-value pairs rather than raw JSON.
	if resp.Result != nil {
		printObject("Result", resp.Result)
	}
}

// printObject prints a nested value under a heading as key-value pairs, its
// fields in name order and each nested value summarized, with a note that
// --json prints all of it. Human output never shows raw JSON.
func printObject(
	title string,
	v any,
) {
	PrintHeading(title)

	m, ok := v.(map[string]any)
	if !ok {
		PrintKV("Value", Summarize(v))
	} else {
		for _, k := range sortedKeys(m) {
			PrintKV(k, Summarize(m[k]))
		}
	}
	PrintNote("--json prints all of it.")
}

// Summarize renders a value for one cell or key-value pair: a scalar as
// itself, a list by its length and an object by its field count, and nothing
// as "none".
func Summarize(
	v any,
) string {
	switch t := v.(type) {
	case nil:
		return "none"
	case []any:
		return Plural(len(t), "item", "items")
	case map[string]any:
		return Plural(len(t), "field", "fields")
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}

		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprint(t)
	}
}

// Plural renders a count with the noun that agrees with it: 1 key, 2 keys.
func Plural(
	n int,
	one string,
	many string,
) string {
	if n == 1 {
		return "1 " + one
	}

	return strconv.Itoa(n) + " " + many
}

// FormatTime renders an RFC 3339 time as local "2006-01-02 15:04:05", the
// one format human output uses. A value that does not parse is returned as
// it is.
func FormatTime(
	s string,
) string {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return s
	}

	return t.Local().Format("2006-01-02 15:04:05")
}

// FormatDuration rounds a Go duration string to what a reader can use:
// 6.168625ms is 6ms, 1.234567s is 1.2s, 184s is 3m4s. A value that does not
// parse is returned as it is.
func FormatDuration(
	s string,
) string {
	d, err := time.ParseDuration(s)
	if err != nil {
		return s
	}

	switch {
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	case d < time.Minute:
		return d.Round(100 * time.Millisecond).String()
	default:
		return d.Round(time.Second).String()
	}
}

// sortedKeys returns a map's keys in order, so maps print the same way every
// run.
func sortedKeys[V any](
	m map[string]V,
) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
