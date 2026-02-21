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

package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/google/uuid"
	"golang.org/x/term"

	"github.com/retr0h/osapi/internal/client/gen"
)

// TODO(retr0h): consider moving out of global scope
var (
	purple    = lipgloss.Color("99")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")
	white     = lipgloss.Color("15")
	teal      = lipgloss.Color("#06ffa5") // Soft teal for values/highlights

	// Reusable inline styles for compact key-value output.
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(purple)
	valueStyle = lipgloss.NewStyle().Foreground(teal)
	dimStyle   = lipgloss.NewStyle().Foreground(gray)
)

// section represents a header with its corresponding rows.
type section struct {
	Title   string
	Headers []string
	Rows    [][]string
}

// resultRow is a per-host broadcast result used by buildBroadcastTable.
type resultRow struct {
	Hostname string
	Error    *string
	Fields   []string
}

// buildBroadcastTable builds headers and rows for a broadcast result table.
// It prepends HOSTNAME to every row and conditionally inserts STATUS and ERROR
// columns when any result carries an error.
func buildBroadcastTable(
	results []resultRow,
	fieldHeaders []string,
) ([]string, [][]string) {
	hasErrors := false
	for _, r := range results {
		if r.Error != nil {
			hasErrors = true
			break
		}
	}

	headers := []string{"HOSTNAME"}
	if hasErrors {
		headers = append(headers, "STATUS", "ERROR")
	}
	headers = append(headers, fieldHeaders...)

	rows := make([][]string, 0, len(results))
	for _, r := range results {
		row := []string{r.Hostname}
		if hasErrors {
			status := "ok"
			errMsg := ""
			if r.Error != nil {
				status = "failed"
				errMsg = *r.Error
			}
			row = append(row, status, errMsg)
		}
		row = append(row, r.Fields...)
		rows = append(rows, row)
	}

	return headers, rows
}

// mutationResultRow is a per-host mutation result used by buildMutationTable.
type mutationResultRow struct {
	Hostname string
	Status   string
	Error    *string
	Fields   []string
}

// buildMutationTable builds headers and rows for a mutation broadcast table.
// Unlike buildBroadcastTable, STATUS and ERROR columns are always shown because
// mutation results carry an explicit status field.
func buildMutationTable(
	results []mutationResultRow,
	fieldHeaders []string,
) ([]string, [][]string) {
	headers := make([]string, 0, 3+len(fieldHeaders))
	headers = append(headers, "HOSTNAME", "STATUS", "ERROR")
	headers = append(headers, fieldHeaders...)

	rows := make([][]string, 0, len(results))
	for _, r := range results {
		errMsg := ""
		if r.Error != nil {
			errMsg = *r.Error
		}
		row := make([]string, 0, 3+len(r.Fields))
		row = append(row, r.Hostname, r.Status, errMsg)
		row = append(row, r.Fields...)
		rows = append(rows, row)
	}

	return headers, rows
}

// printStyledTable renders a styled table with dynamic column widths.
func printStyledTable(
	sections []section,
) {
	re := lipgloss.NewRenderer(os.Stdout)

	// Get terminal width to constrain table if needed
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 120 // Default to reasonable width if unable to get terminal size
	}

	for _, section := range sections {
		// Calculate optimal width for each column
		columnWidths := calculateColumnWidths(section.Headers, section.Rows, 1)

		// Check if total table width exceeds terminal width
		totalWidth := 0
		for _, width := range columnWidths {
			totalWidth += width
		}

		// Add border/spacing overhead (rough estimate)
		totalWidth += len(columnWidths) * 3 // borders and spacing

		// If table is too wide, proportionally reduce column widths
		if totalWidth > termWidth-4 { // leave some margin
			scale := float64(termWidth-4) / float64(totalWidth)
			for i := range columnWidths {
				columnWidths[i] = int(float64(columnWidths[i]) * scale)
				if columnWidths[i] < 8 { // minimum readable width
					columnWidths[i] = 8
				}
			}
		}

		var (
			HeaderStyle  = re.NewStyle().Foreground(white).Bold(true).Align(lipgloss.Center)
			CellStyle    = re.NewStyle().PaddingLeft(1)
			OddRowStyle  = CellStyle.Foreground(gray)
			EvenRowStyle = CellStyle.Foreground(lightGray)
			BorderStyle  = re.NewStyle().Foreground(purple)
			PaddingStyle = re.NewStyle().Padding(0, 2)
			TitleStyle   = re.NewStyle().Bold(true).Foreground(purple).PaddingLeft(2).PaddingTop(1)
			ColonStyle   = re.NewStyle().Bold(false)
		)

		if section.Title != "" {
			titleWithColon := TitleStyle.Render(section.Title) + ColonStyle.Render(":")
			fmt.Println(titleWithColon)
		} else {
			fmt.Println()
		}

		// Create the table and apply styles.
		t := table.New().
			Border(lipgloss.ThickBorder()).
			BorderStyle(BorderStyle).
			StyleFunc(func(
				row int,
				col int,
			) lipgloss.Style {
				// Determine base style based on row
				var baseStyle lipgloss.Style
				switch row % 2 {
				case 0:
					baseStyle = EvenRowStyle
				default:
					baseStyle = OddRowStyle
				}

				// Apply column-specific width if available
				if col < len(columnWidths) {
					baseStyle = baseStyle.Width(columnWidths[col])
				}

				return baseStyle
			})

		styledHeaders := make([]string, len(section.Headers))
		for i, header := range section.Headers {
			styledHeaders[i] = HeaderStyle.Render(header)
		}
		t.Headers(styledHeaders...)

		t.Rows(section.Rows...)

		// Render the styled table.
		fmt.Println(PaddingStyle.Render(t.String()))
	}
}

// kvMinColWidth is the minimum visual width for each key-value column.
// A consistent minimum ensures columns align across consecutive printKV calls.
const kvMinColWidth = 20

// printKV prints labeled key-value pairs on a single indented line.
// Pairs are padded to equal column widths for alignment.
// Arguments alternate between labels and values: label1, val1, label2, val2, ...
func printKV(
	pairs ...string,
) {
	if len(pairs)%2 != 0 || len(pairs) == 0 {
		return
	}

	rendered := make([]string, 0, len(pairs)/2)
	maxWidth := kvMinColWidth
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

// formatList helper function to convert []string to a formatted string.
func formatList(
	list []string,
) string {
	if len(list) == 0 {
		return "None"
	}
	return strings.Join(list, ", ")
}

// calculateColumnWidths calculates the optimal width for each column based on content
func calculateColumnWidths(
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
				maxLineWidth := getMaxLineWidth(cell)
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

// getMaxLineWidth returns the width of the longest line in a multi-line string
func getMaxLineWidth(
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

// safeString function to safely dereference string pointers
func safeString(
	s *string,
) string {
	if s != nil {
		return *s
	}
	return ""
}

// safeUUID converts a *uuid.UUID to its string representation. Returns "" if nil.
func safeUUID(
	u *uuid.UUID,
) string {
	if u != nil {
		return u.String()
	}
	return ""
}

// float64ToSafeString converts a *float64 to a string. Returns "N/A" if nil.
func float64ToSafeString(
	f *float64,
) string {
	if f != nil {
		return fmt.Sprintf("%f", *f)
	}
	return "N/A"
}

// intToSafeString converts a *int to a string. Returns "N/A" if nil.
func intToSafeString(
	i *int,
) string {
	if i != nil {
		return fmt.Sprintf("%d", *i)
	}
	return "N/A"
}

// handleAuthError handles authentication and authorization errors (401 and 403).
func handleAuthError(
	jsonError *gen.ErrorResponse,
	statusCode int,
	logger *slog.Logger,
) {
	errorMsg := "unknown error"

	if jsonError != nil && jsonError.Error != nil {
		errorMsg = safeString(jsonError.Error)
	}

	logger.Error(
		"authorization error",
		slog.Int("code", statusCode),
		slog.String("response", errorMsg),
	)
}

// handleUnknownError handles unexpected errors, such as 500 Internal Server Error.
func handleUnknownError(
	json500 *gen.ErrorResponse,
	statusCode int,
	logger *slog.Logger,
) {
	errorMsg := "unknown error"

	if json500 != nil && json500.Error != nil {
		errorMsg = safeString(json500.Error)
	}

	logger.Error(
		"error in response",
		slog.Int("code", statusCode),
		slog.String("error", errorMsg),
	)
}

// displayJobDetailResponse displays detailed job information from a REST API response.
// Used by both job get and job run commands.
func displayJobDetailResponse(
	resp *gen.JobDetailResponse,
) {
	// Display job metadata
	fmt.Println()
	printKV("Job ID", safeUUID(resp.Id), "Status", safeString(resp.Status))
	if resp.Hostname != nil && *resp.Hostname != "" {
		printKV("Hostname", *resp.Hostname)
	}
	if resp.Created != nil {
		printKV("Created", *resp.Created)
	}
	if resp.UpdatedAt != nil && *resp.UpdatedAt != "" {
		printKV("Updated At", *resp.UpdatedAt)
	}
	if resp.Error != nil && *resp.Error != "" {
		printKV("Error", *resp.Error)
	}

	// Add worker summary from worker_states
	if resp.WorkerStates != nil && len(*resp.WorkerStates) > 0 {
		completed := 0
		failed := 0
		processing := 0

		for _, state := range *resp.WorkerStates {
			if state.Status != nil {
				switch *state.Status {
				case "completed":
					completed++
				case "failed":
					failed++
				case "started":
					processing++
				}
			}
		}

		total := len(*resp.WorkerStates)
		if total > 1 {
			printKV("Workers", fmt.Sprintf(
				"%d total (%d completed, %d failed, %d processing)",
				total,
				completed,
				failed,
				processing,
			))
		}
	}

	var sections []section

	// Display the operation request
	if resp.Operation != nil {
		jobOperationJSON, _ := json.MarshalIndent(*resp.Operation, "", "  ")
		operationRows := [][]string{{string(jobOperationJSON)}}
		sections = append(sections, section{
			Title:   "Job Request",
			Headers: []string{"DATA"},
			Rows:    operationRows,
		})
	}

	// Display worker responses (for broadcast jobs)
	if resp.Responses != nil && len(*resp.Responses) > 0 {
		responseRows := make([][]string, 0, len(*resp.Responses))
		for hostname, response := range *resp.Responses {
			status := safeString(response.Status)
			errMsg := ""
			if response.Error != nil {
				errMsg = *response.Error
			}

			var dataStr string
			if response.Data != nil {
				dataJSON, _ := json.MarshalIndent(response.Data, "", "  ")
				dataStr = string(dataJSON)
			} else {
				dataStr = "(no data)"
			}

			row := []string{hostname, status, dataStr, errMsg}
			responseRows = append(responseRows, row)
		}

		sections = append(sections, section{
			Title:   "Worker Responses",
			Headers: []string{"HOSTNAME", "STATUS", "DATA", "ERROR"},
			Rows:    responseRows,
		})
	}

	// Display worker states (for broadcast jobs)
	if resp.WorkerStates != nil && len(*resp.WorkerStates) > 0 {
		stateRows := make([][]string, 0, len(*resp.WorkerStates))
		for hostname, state := range *resp.WorkerStates {
			status := safeString(state.Status)
			duration := safeString(state.Duration)
			errMsg := ""
			if state.Error != nil {
				errMsg = *state.Error
			}

			stateRows = append(stateRows, []string{hostname, status, duration, errMsg})
		}

		sections = append(sections, section{
			Title:   "Worker States",
			Headers: []string{"HOSTNAME", "STATUS", "DURATION", "ERROR"},
			Rows:    stateRows,
		})
	}

	// Display timeline
	if resp.Timeline != nil && len(*resp.Timeline) > 0 {
		timelineRows := make([][]string, 0, len(*resp.Timeline))
		for _, te := range *resp.Timeline {
			ts := safeString(te.Timestamp)
			event := safeString(te.Event)
			hostname := safeString(te.Hostname)
			message := safeString(te.Message)
			errMsg := ""
			if te.Error != nil {
				errMsg = *te.Error
			}
			timelineRows = append(timelineRows, []string{ts, event, hostname, message, errMsg})
		}

		sections = append(sections, section{
			Title:   "Timeline",
			Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
			Rows:    timelineRows,
		})
	}

	// Display result if completed
	if resp.Result != nil {
		resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
		resultRows := [][]string{{string(resultJSON)}}
		sections = append(sections, section{
			Title:   "Job Result",
			Headers: []string{"DATA"},
			Rows:    resultRows,
		})
	}

	for _, sec := range sections {
		printStyledTable([]section{sec})
	}
}
