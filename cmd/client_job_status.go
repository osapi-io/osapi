// Copyright (c) 2025 John Dewey

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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type jobsModel struct {
	jobsStatus       string
	lastUpdate       time.Time
	isLoading        bool
	pollIntervalSecs int
}

func initialJobsModel(pollIntervalSecs int) jobsModel {
	return jobsModel{
		jobsStatus:       "Fetching jobs status...",
		lastUpdate:       time.Now(),
		isLoading:        true,
		pollIntervalSecs: pollIntervalSecs,
	}
}

func (m jobsModel) tickCmd() tea.Cmd {
	pollInterval := time.Duration(m.pollIntervalSecs) * time.Second

	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return t
	})
}

func fetchJobsCmd() tea.Cmd {
	return func() tea.Msg {
		return fetchJobsStatus()
	}
}

func (m jobsModel) Init() tea.Cmd {
	return tea.Batch(fetchJobsCmd(), m.tickCmd())
}

func (m jobsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case string:
		m.jobsStatus = msg
		m.lastUpdate = time.Now()
		m.isLoading = false
		return m, m.tickCmd()
	case time.Time:
		// timer ticks, fetch new jobs status
		return m, fetchJobsCmd()
	}
	return m, nil
}

func (m jobsModel) View() string {
	var (
		titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(purple)
		timeStyle   = lipgloss.NewStyle().Foreground(lightGray).Italic(true)
		borderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1).
				Margin(2).
				BorderForeground(purple)
	)

	title := titleStyle.Render("Jobs Queue Status")

	// Apply styling to the status text with colored keys/values
	styledStatus := styleStatusText(m.jobsStatus)

	lastUpdated := timeStyle.Render(
		fmt.Sprintf("Last Updated: %v", m.lastUpdate.Format(time.RFC1123)),
	)
	quitInstruction := timeStyle.Render("Press 'q' to quit")

	return borderStyle.Render(
		fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, styledStatus, lastUpdated, quitInstruction),
	)
}

func styleStatusText(statusText string) string {
	var (
		keyStyle     = lipgloss.NewStyle().Foreground(gray) // Gray for all keys
		valueStyle   = lipgloss.NewStyle().Foreground(teal) // Soft teal for values
		sectionStyle = lipgloss.NewStyle().Foreground(gray) // Gray for section headers
	)

	lines := strings.Split(statusText, "\n")
	var styledLines []string

	for _, line := range lines {
		// Check if this is a section header (no indentation, ends with colon, no value after)
		if strings.HasSuffix(strings.TrimSpace(line), ":") && !strings.HasPrefix(line, "  ") {
			// This is a section header like "Jobs Queue Status:" or "Operation Types:"
			styledLines = append(styledLines, sectionStyle.Render(line))
		} else if strings.Contains(line, ":") {
			// Split on the first colon to separate key and value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Preserve indentation
				indent := ""
				if strings.HasPrefix(line, "  ") {
					indent = "  "
				}

				styledLine := indent + keyStyle.Render(key+":") + " " + valueStyle.Render(value)
				styledLines = append(styledLines, styledLine)
			} else {
				styledLines = append(styledLines, line)
			}
		} else {
			// No colon, just render as is
			styledLines = append(styledLines, line)
		}
	}

	return strings.Join(styledLines, "\n")
}

func fetchJobsStatus() string {
	stats, err := jobClient.GetQueueStats(context.Background())
	if err != nil {
		return fmt.Sprintf("Error fetching jobs: %v", err)
	}

	if stats.TotalJobs == 0 {
		return "Job queue is empty (0 jobs total)"
	}

	// Build status display
	statusDisplay := "Jobs Queue Status:\n"
	statusDisplay += fmt.Sprintf("  Total Jobs: %d\n", stats.TotalJobs)
	statusDisplay += fmt.Sprintf("  Unprocessed: %d\n", stats.StatusCounts["unprocessed"])
	statusDisplay += fmt.Sprintf("  Processing: %d\n", stats.StatusCounts["processing"])
	statusDisplay += fmt.Sprintf("  Completed: %d\n", stats.StatusCounts["completed"])
	statusDisplay += fmt.Sprintf("  Failed: %d\n", stats.StatusCounts["failed"])
	if stats.DLQCount > 0 {
		statusDisplay += fmt.Sprintf("  Dead Letter Queue: %d\n", stats.DLQCount)
	}

	if len(stats.OperationCounts) > 0 {
		statusDisplay += "\nOperation Types:\n"
		for opType, count := range stats.OperationCounts {
			statusDisplay += fmt.Sprintf("  %s: %d\n", opType, count)
		}
	}

	return statusDisplay
}

func fetchJobsStatusJSON() string {
	stats, err := jobClient.GetQueueStats(context.Background())
	if err != nil {
		errorResult := map[string]interface{}{
			"error": fmt.Sprintf("Error fetching jobs: %v", err),
		}
		resultJSON, _ := json.Marshal(errorResult)
		return string(resultJSON)
	}

	resultJSON, _ := json.Marshal(stats)
	return string(resultJSON)
}

// clientJobStatusCmd represents the clientJobsStatus command.
var clientJobStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the jobs queue status",
	Long: `Displays the jobs queue status with automatic updates.
Shows job counts by status (unprocessed/processing/completed/failed) 
and operation types with live refresh.`,
	Run: func(cmd *cobra.Command, _ []string) {
		pollIntervalSeconds, _ := cmd.Flags().GetInt("poll-interval-seconds")

		// Check if running in non-interactive mode (JSON output or no TTY)
		if jsonOutput {
			// Get status once and output as JSON
			status := fetchJobsStatusJSON()
			logger.Info("jobs status", slog.String("response", status))
			return
		}

		// Run interactive TUI
		p := tea.NewProgram(initialJobsModel(pollIntervalSeconds))
		_, err := p.Run()
		if err != nil {
			// Fallback to non-interactive mode if TUI fails
			status := fetchJobsStatus()
			fmt.Println(status)
		}
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobStatusCmd)

	clientJobStatusCmd.PersistentFlags().
		Int("poll-interval-seconds", 30, "The interval (in seconds) between polling operations")
}
