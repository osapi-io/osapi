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
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

type jobsModel struct {
	jobsStatus       string
	lastUpdate       time.Time
	isLoading        bool
	pollIntervalSecs int
}

func initialJobsModel(
	pollIntervalSecs int,
) jobsModel {
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

func (m jobsModel) Update(
	msg tea.Msg,
) (tea.Model, tea.Cmd) {
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
		return m, fetchJobsCmd()
	}
	return m, nil
}

func (m jobsModel) View() string {
	var (
		titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(cli.Purple)
		timeStyle   = lipgloss.NewStyle().Foreground(cli.LightGray).Italic(true)
		borderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1).
				Margin(2).
				BorderForeground(cli.Purple)
	)

	title := titleStyle.Render("Jobs Queue Status")

	styledStatus := styleStatusText(m.jobsStatus)

	lastUpdated := timeStyle.Render(
		fmt.Sprintf("Last Updated: %v", m.lastUpdate.Format(time.RFC1123)),
	)
	quitInstruction := timeStyle.Render("Press 'q' to quit")

	return borderStyle.Render(
		fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, styledStatus, lastUpdated, quitInstruction),
	)
}

func styleStatusText(
	statusText string,
) string {
	var (
		keyStyle     = lipgloss.NewStyle().Foreground(cli.Gray)
		valueStyle   = lipgloss.NewStyle().Foreground(cli.Teal)
		sectionStyle = lipgloss.NewStyle().Foreground(cli.Gray)
	)

	lines := strings.Split(statusText, "\n")
	var styledLines []string

	for _, line := range lines {
		if strings.HasSuffix(strings.TrimSpace(line), ":") && !strings.HasPrefix(line, "  ") {
			styledLines = append(styledLines, sectionStyle.Render(line))
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

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
			styledLines = append(styledLines, line)
		}
	}

	return strings.Join(styledLines, "\n")
}

func fetchJobsStatus() string {
	resp, err := sdkClient.Job.QueueStats(context.Background())
	if err != nil {
		return fmt.Sprintf("Error fetching jobs: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Sprintf("Error fetching jobs: HTTP %d", resp.StatusCode())
	}

	stats := resp.JSON200
	if stats == nil {
		return "Error fetching jobs: nil response"
	}

	totalJobs := 0
	if stats.TotalJobs != nil {
		totalJobs = *stats.TotalJobs
	}

	if totalJobs == 0 {
		return "Job queue is empty (0 jobs total)"
	}

	statusCounts := map[string]int{}
	if stats.StatusCounts != nil {
		statusCounts = *stats.StatusCounts
	}

	statusDisplay := "Jobs Queue Status:\n"
	statusDisplay += fmt.Sprintf("  Total Jobs: %d\n", totalJobs)
	statusDisplay += fmt.Sprintf("  Unprocessed: %d\n", statusCounts["unprocessed"])
	statusDisplay += fmt.Sprintf("  Processing: %d\n", statusCounts["processing"])
	statusDisplay += fmt.Sprintf("  Completed: %d\n", statusCounts["completed"])
	statusDisplay += fmt.Sprintf("  Failed: %d\n", statusCounts["failed"])

	if stats.DlqCount != nil && *stats.DlqCount > 0 {
		statusDisplay += fmt.Sprintf("  Dead Letter Queue: %d\n", *stats.DlqCount)
	}

	if stats.OperationCounts != nil && len(*stats.OperationCounts) > 0 {
		statusDisplay += "\nOperation Types:\n"
		for opType, count := range *stats.OperationCounts {
			statusDisplay += fmt.Sprintf("  %s: %d\n", opType, count)
		}
	}

	return statusDisplay
}

func fetchJobsStatusJSON() string {
	resp, err := sdkClient.Job.QueueStats(context.Background())
	if err != nil {
		errorResult := map[string]interface{}{
			"error": fmt.Sprintf("Error fetching jobs: %v", err),
		}
		resultJSON, _ := json.Marshal(errorResult)
		return string(resultJSON)
	}

	if resp.StatusCode() != http.StatusOK {
		errorResult := map[string]interface{}{
			"error": fmt.Sprintf("HTTP %d", resp.StatusCode()),
		}
		resultJSON, _ := json.Marshal(errorResult)
		return string(resultJSON)
	}

	return string(resp.Body)
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

		if jsonOutput {
			status := fetchJobsStatusJSON()
			fmt.Println(status)
			return
		}

		p := tea.NewProgram(initialJobsModel(pollIntervalSeconds))
		_, err := p.Run()
		if err != nil {
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
