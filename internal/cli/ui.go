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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// Theme colors for terminal UI rendering.
var (
	Purple    = lipgloss.Color("99")
	Gray      = lipgloss.Color("245")
	LightGray = lipgloss.Color("241")
	White     = lipgloss.Color("15")
	Teal      = lipgloss.Color("#06ffa5")
)

// Reusable inline styles for compact key-value output.
var (
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(Purple)
	valueStyle = lipgloss.NewStyle().Foreground(Teal)

	// DimStyle is a muted style for secondary text.
	DimStyle = lipgloss.NewStyle().Foreground(Gray)
)

// Section represents a header with its corresponding rows.
type Section struct {
	Title   string
	Headers []string
	Rows    [][]string
}

// ResultRow is a per-host broadcast result used by BuildBroadcastTable.
type ResultRow struct {
	Hostname string
	Status   string
	Changed  *bool
	Error    *string
	Fields   []string
}

// BuildBroadcastTable builds headers and rows for a broadcast result table.
// It prepends HOSTNAME to every row and conditionally inserts STATUS, CHANGED,
// and ERROR columns when any result carries an error.
func BuildBroadcastTable(
	results []ResultRow,
	fieldHeaders []string,
) ([]string, [][]string) {
	hasErrors := false
	for _, r := range results {
		if r.Error != nil {
			hasErrors = true
			break
		}
	}

	hasChanged := false
	for _, r := range results {
		if r.Changed != nil {
			hasChanged = true
			break
		}
	}

	headers := []string{"HOSTNAME"}
	if hasErrors {
		headers = append(headers, "STATUS", "ERROR")
	}
	if hasChanged {
		headers = append(headers, "CHANGED")
	}
	headers = append(headers, fieldHeaders...)

	rows := make([][]string, 0, len(results))
	for _, r := range results {
		row := []string{r.Hostname}
		if hasErrors {
			status := r.Status
			if status == "" {
				status = "ok"
			}
			errMsg := ""
			if r.Error != nil {
				errMsg = *r.Error
			}
			row = append(row, status, errMsg)
		}
		if hasChanged {
			changedStr := ""
			if r.Changed != nil {
				changedStr = fmt.Sprintf("%v", *r.Changed)
			}
			row = append(row, changedStr)
		}
		row = append(row, r.Fields...)
		rows = append(rows, row)
	}

	return headers, rows
}

// MutationResultRow is a per-host mutation result used by BuildMutationTable.
type MutationResultRow struct {
	Hostname string
	Status   string
	Changed  *bool
	Error    *string
	Fields   []string
}

// BuildMutationTable builds headers and rows for a mutation broadcast table.
// Unlike BuildBroadcastTable, STATUS and ERROR columns are always shown because
// mutation results carry an explicit status field.
func BuildMutationTable(
	results []MutationResultRow,
	fieldHeaders []string,
) ([]string, [][]string) {
	headers := make([]string, 0, 4+len(fieldHeaders))
	headers = append(headers, "HOSTNAME", "STATUS", "CHANGED", "ERROR")
	headers = append(headers, fieldHeaders...)

	rows := make([][]string, 0, len(results))
	for _, r := range results {
		errMsg := ""
		if r.Error != nil {
			errMsg = *r.Error
		}
		changedStr := ""
		if r.Changed != nil {
			changedStr = fmt.Sprintf("%v", *r.Changed)
		}
		row := make([]string, 0, 4+len(r.Fields))
		row = append(row, r.Hostname, r.Status, changedStr, errMsg)
		row = append(row, r.Fields...)
		rows = append(rows, row)
	}

	return headers, rows
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

// compactMaxColWidth is the maximum column width before truncation.
const compactMaxColWidth = 79

// printJSONBlock prints a titled JSON block without truncation.
func printJSONBlock(
	title string,
	jsonStr string,
) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(Purple)
	dataStyle := lipgloss.NewStyle().Foreground(Teal)

	fmt.Printf("\n  %s:\n", titleStyle.Render(title))
	fmt.Printf("  %s\n", dataStyle.Render(jsonStr))
}

// PrintCompactTable renders a compact column-aligned table (kubectl-style).
// Headers are uppercase purple, data rows are teal, with 2-space indent.
// Multi-line cell values are flattened to a single line and long values
// are truncated with an ellipsis.
func PrintCompactTable(
	sections []Section,
) {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(Purple)
	evenStyle := lipgloss.NewStyle().Foreground(Teal)
	oddStyle := lipgloss.NewStyle().Foreground(White)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(Purple)

	const colGap = 2

	for _, section := range sections {
		if section.Title != "" {
			fmt.Printf("\n  %s:\n", titleStyle.Render(section.Title))
		} else {
			fmt.Println()
		}

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

		// Build data rows with alternating colors
		for r, row := range flatRows {
			rowStyle := evenStyle
			if r%2 != 0 {
				rowStyle = oddStyle
			}
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
				if i < len(section.Headers)-1 {
					line.WriteString(rowStyle.Render(fmt.Sprintf("%-*s", widths[i]+colGap, cell)))
				} else {
					line.WriteString(rowStyle.Render(cell))
				}
			}
			fmt.Println(line.String())
		}
	}
}

// KVMinColWidth is the minimum visual width for each key-value column.
// A consistent minimum ensures columns align across consecutive PrintKV calls.
const KVMinColWidth = 20

// PrintKV prints labeled key-value pairs on a single indented line.
// Pairs are padded to equal column widths for alignment.
// Arguments alternate between labels and values: label1, val1, label2, val2, ...
func PrintKV(
	pairs ...string,
) {
	if len(pairs)%2 != 0 || len(pairs) == 0 {
		return
	}

	rendered := make([]string, 0, len(pairs)/2)
	maxWidth := KVMinColWidth
	for i := 0; i < len(pairs); i += 2 {
		pair := labelStyle.Render(pairs[i]+":") + " " + valueStyle.Render(pairs[i+1])
		rendered = append(rendered, pair)
		if w := lipgloss.Width(pair); w > maxWidth {
			maxWidth = w
		}
	}

	var line strings.Builder
	line.WriteString("  ")
	for i, pair := range rendered {
		line.WriteString(pair)
		if i < len(rendered)-1 {
			pad := maxWidth - lipgloss.Width(pair) + 4
			line.WriteString(strings.Repeat(" ", pad))
		}
	}
	fmt.Println(line.String())
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

// HandleError logs the appropriate error message based on the SDK error type.
func HandleError(
	err error,
	logger *slog.Logger,
) {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		logger.Error(
			"api error",
			slog.Int("code", apiErr.StatusCode),
			slog.String("error", apiErr.Message),
		)

		return
	}

	logger.Error("unexpected error", slog.String("error", err.Error()))
}

// DisplayJobDetail displays detailed job information from domain types.
// Used by both job get and job run commands.
func DisplayJobDetail(
	resp *client.JobDetail,
) {
	// Display job metadata
	fmt.Println()
	PrintKV("Job ID", resp.ID, "Status", resp.Status)
	if resp.Hostname != "" {
		PrintKV("Hostname", resp.Hostname)
	}
	if resp.Created != "" {
		PrintKV("Created", resp.Created)
	}
	if resp.UpdatedAt != "" {
		PrintKV("Updated At", resp.UpdatedAt)
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

	// Display the operation request as an untruncated JSON block
	if resp.Operation != nil {
		jobOperationJSON, _ := json.MarshalIndent(resp.Operation, "  ", "  ")
		printJSONBlock("Job Request", string(jobOperationJSON))
	}

	var sections []Section

	// Display agent responses (for broadcast jobs)
	if len(resp.Responses) > 0 {
		responseRows := make([][]string, 0, len(resp.Responses))
		for hostname, response := range resp.Responses {
			var dataStr string
			if response.Data != nil {
				dataJSON, _ := json.MarshalIndent(response.Data, "", "  ")
				dataStr = string(dataJSON)
			} else {
				dataStr = "(no data)"
			}

			row := []string{hostname, response.Status, dataStr, response.Error}
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
		for hostname, state := range resp.AgentStates {
			stateRows = append(
				stateRows,
				[]string{hostname, state.Status, state.Duration, state.Error},
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
	if len(timelineRows) == 0 {
		timelineRows = [][]string{{"No events"}}
	}
	sections = append(sections, Section{
		Title:   "Timeline",
		Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
		Rows:    timelineRows,
	})

	for _, sec := range sections {
		PrintCompactTable([]Section{sec})
	}

	// Display result as an untruncated JSON block after tables
	if resp.Result != nil {
		resultJSON, _ := json.MarshalIndent(resp.Result, "  ", "  ")
		printJSONBlock("Job Result", string(resultJSON))
	}
}
