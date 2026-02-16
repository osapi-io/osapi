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
	"golang.org/x/term"

	"github.com/retr0h/osapi/internal/client/gen"
	"github.com/retr0h/osapi/internal/job"
)

// TODO(retr0h): consider moving out of global scope
var (
	purple    = lipgloss.Color("99")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")
	white     = lipgloss.Color("15")
	teal      = lipgloss.Color("#06ffa5") // Soft teal for values/highlights
)

// section represents a header with its corresponding rows.
type section struct {
	Title   string
	Headers []string
	Rows    [][]string
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
			TitleStyle   = re.NewStyle().Bold(true).Foreground(purple).PaddingLeft(2).PaddingTop(2)
			ColonStyle   = re.NewStyle().Bold(false).MarginBottom(1)
		)

		if section.Title != "" {
			titleWithColon := TitleStyle.Render(section.Title) + ColonStyle.Render(":")
			fmt.Println(titleWithColon)
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

// printStyledMap format and print the map into a styled, padded table.
func printStyledMap(
	data map[string]interface{},
) {
	paddingStyle := lipgloss.NewStyle().Padding(1, 2)

	var builder strings.Builder

	for key, value := range data {
		styledKey := lipgloss.NewStyle().
			Bold(true).
			Foreground(purple).
			Render(key)

		formattedLine := fmt.Sprintf("\n%s: %v", styledKey, value)
		builder.WriteString(formattedLine)
	}

	paddedOutput := paddingStyle.Render(builder.String())

	fmt.Println(paddedOutput)
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

// displayJobDetails displays detailed job information in a consistent format.
// Used by both job get and job run commands.
func displayJobDetails(
	jobInfo *job.QueuedJob,
) {
	if jsonOutput {
		resultJSON, _ := json.Marshal(jobInfo)
		logger.Info("job", slog.String("response", string(resultJSON)))
		return
	}

	// Display job details with enhanced status information
	jobData := map[string]interface{}{
		"Job ID":  jobInfo.ID,
		"Status":  jobInfo.Status,
		"Created": jobInfo.Created,
	}

	// Add error field if present
	if jobInfo.Error != "" {
		jobData["Error"] = jobInfo.Error
	}

	// Add subject if present
	if jobInfo.Subject != "" {
		jobData["Subject"] = jobInfo.Subject
	}

	// Add worker summary
	if len(jobInfo.WorkerStates) > 0 {
		completed := 0
		failed := 0
		processing := 0
		acknowledged := 0

		for _, state := range jobInfo.WorkerStates {
			switch state.Status {
			case "completed":
				completed++
			case "failed":
				failed++
			case "started":
				processing++
			case "acknowledged":
				acknowledged++
			}
		}

		total := len(jobInfo.WorkerStates)
		if total > 1 {
			jobData["Workers"] = fmt.Sprintf(
				"%d total (%d completed, %d failed, %d processing)",
				total,
				completed,
				failed,
				processing,
			)
		}
	}

	printStyledMap(jobData)

	// Collect content for sections to ensure consistent table widths
	var sections []section

	// Display the operation request
	if jobInfo.Operation != nil {
		jobOperationJSON, _ := json.MarshalIndent(jobInfo.Operation, "", "  ")
		operationRows := [][]string{{string(jobOperationJSON)}}
		sections = append(sections, section{
			Title:   "Job Request",
			Headers: []string{"DATA"},
			Rows:    operationRows,
		})
	}

	// Display timeline if available
	if len(jobInfo.Timeline) > 0 {
		timelineRows := [][]string{}
		for _, event := range jobInfo.Timeline {
			row := []string{
				event.Timestamp.Format("15:04:05 MST"),
				event.Event,
				event.Hostname,
				event.Message,
			}
			if event.Error != "" {
				row = append(row, event.Error)
			} else {
				row = append(row, "")
			}
			timelineRows = append(timelineRows, row)
		}
		sections = append(sections, section{
			Title:   "Timeline",
			Headers: []string{"TIME", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
			Rows:    timelineRows,
		})
	}

	// Display responses (the actual job results)
	if len(jobInfo.Responses) > 0 {
		responseRows := [][]string{}
		for hostname, response := range jobInfo.Responses {
			// Format the response data with nice indentation
			var responseData string
			if len(response.Data) > 0 {
				var data interface{}
				if err := json.Unmarshal(response.Data, &data); err == nil {
					// Use pretty printing with 2-space indentation
					dataJSON, _ := json.MarshalIndent(data, "", "  ")
					responseData = string(dataJSON)
				} else {
					responseData = string(response.Data)
				}
			} else {
				responseData = "(no data)"
			}

			row := []string{
				hostname,
				string(response.Status),
				response.Timestamp.Format("15:04:05 MST"),
				responseData,
			}
			if response.Error != "" {
				row = append(row, response.Error)
			} else {
				row = append(row, "")
			}
			responseRows = append(responseRows, row)
		}

		sections = append(sections, section{
			Title:   "Worker Responses",
			Headers: []string{"HOSTNAME", "STATUS", "TIME", "DATA", "ERROR"},
			Rows:    responseRows,
		})
	}

	// Display results if completed and available (legacy support)
	if jobInfo.Status == "completed" && len(jobInfo.Result) > 0 {
		var result interface{}
		if err := json.Unmarshal(jobInfo.Result, &result); err == nil {
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			resultRows := [][]string{{string(resultJSON)}}
			sections = append(sections, section{
				Title:   "Job Response (Legacy)",
				Headers: []string{"DATA"},
				Rows:    resultRows,
			})
		}
	}

	// Print each section individually to ensure consistent formatting
	for _, sec := range sections {
		printStyledTable([]section{sec})
	}
}
