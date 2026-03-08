package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
)

type RunnerBroadcastTestSuite struct {
	suite.Suite
}

func TestRunnerBroadcastTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerBroadcastTestSuite))
}

func (s *RunnerBroadcastTestSuite) TestIsBroadcastTarget() {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "all agents is broadcast",
			target: "_all",
			want:   true,
		},
		{
			name:   "label selector is broadcast",
			target: "role:web",
			want:   true,
		},
		{
			name:   "single agent is not broadcast",
			target: "agent-001",
			want:   false,
		},
		{
			name:   "empty string is not broadcast",
			target: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := IsBroadcastTarget(tt.target)
			s.Equal(tt.want, got)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestExtractHostResults() {
	tests := []struct {
		name string
		data map[string]any
		want []HostResult
	}{
		{
			name: "extracts host results from results array",
			data: map[string]any{
				"results": []any{
					map[string]any{
						"hostname": "host-1",
						"changed":  true,
						"data":     "something",
					},
					map[string]any{
						"hostname": "host-2",
						"changed":  false,
						"error":    "connection refused",
					},
				},
			},
			want: []HostResult{
				{
					Hostname: "host-1",
					Changed:  true,
					Data: map[string]any{
						"hostname": "host-1",
						"changed":  true,
						"data":     "something",
					},
				},
				{
					Hostname: "host-2",
					Changed:  false,
					Error:    "connection refused",
					Data: map[string]any{
						"hostname": "host-2",
						"changed":  false,
						"error":    "connection refused",
					},
				},
			},
		},
		{
			name: "no results key returns nil",
			data: map[string]any{
				"other": "value",
			},
			want: nil,
		},
		{
			name: "results not an array returns nil",
			data: map[string]any{
				"results": "not-an-array",
			},
			want: nil,
		},
		{
			name: "empty results array returns empty slice",
			data: map[string]any{
				"results": []any{},
			},
			want: []HostResult{},
		},
		{
			name: "non-map item in results array is skipped",
			data: map[string]any{
				"results": []any{
					"not-a-map",
					42,
					map[string]any{
						"hostname": "host-1",
						"changed":  true,
					},
				},
			},
			want: []HostResult{
				{
					Hostname: "host-1",
					Changed:  true,
					Data: map[string]any{
						"hostname": "host-1",
						"changed":  true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := extractHostResults(tt.data)
			s.Equal(tt.want, got)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestIsCommandOp() {
	tests := []struct {
		name      string
		operation string
		want      bool
	}{
		{
			name:      "command.exec.execute is a command op",
			operation: "command.exec.execute",
			want:      true,
		},
		{
			name:      "command.shell.execute is a command op",
			operation: "command.shell.execute",
			want:      true,
		},
		{
			name:      "node.hostname.get is not a command op",
			operation: "node.hostname.get",
			want:      false,
		},
		{
			name:      "empty string is not a command op",
			operation: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := isCommandOp(tt.operation)
			s.Equal(tt.want, got)
		})
	}
}

// jobTestServer creates an httptest server that handles POST /job
// and GET /job/{id} with the provided result payload.
func jobTestServer(
	jobResult map[string]any,
) *httptest.Server {
	const jobID = "11111111-1111-1111-1111-111111111111"

	return httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			resp := map[string]any{
				"job_id": jobID,
				"status": "created",
			}
			_ = json.NewEncoder(w).Encode(resp)

		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/job/%s", jobID):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			resp := map[string]any{
				"id":     jobID,
				"status": "completed",
				"result": jobResult,
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func (s *RunnerBroadcastTestSuite) TestExecuteOpPopulatesJobID() {
	tests := []struct {
		name      string
		jobResult map[string]any
		wantJobID string
	}{
		{
			name:      "op result carries job ID",
			jobResult: map[string]any{"hostname": "web-01"},
			wantJobID: "11111111-1111-1111-1111-111111111111",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			origInterval := DefaultPollInterval
			DefaultPollInterval = 10 * time.Millisecond

			defer func() {
				DefaultPollInterval = origInterval
			}()

			srv := jobTestServer(tt.jobResult)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(StopAll))

			plan.Task("get-hostname", &Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			})

			report, err := plan.Run(context.Background())

			s.Require().NoError(err)
			s.Require().Len(report.Tasks, 1)
			s.Equal(tt.wantJobID, report.Tasks[0].JobID)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestExecuteOpBroadcast() {
	tests := []struct {
		name            string
		jobResult       map[string]any
		wantHostResults int
		wantHostname    string
	}{
		{
			name: "broadcast op extracts host results",
			jobResult: map[string]any{
				"results": []any{
					map[string]any{
						"hostname": "host-1",
						"changed":  true,
					},
					map[string]any{
						"hostname": "host-2",
						"changed":  false,
					},
				},
			},
			wantHostResults: 2,
			wantHostname:    "host-1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			origInterval := DefaultPollInterval
			DefaultPollInterval = 10 * time.Millisecond

			defer func() {
				DefaultPollInterval = origInterval
			}()

			srv := jobTestServer(tt.jobResult)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(StopAll))

			plan.Task("broadcast-op", &Op{
				Operation: "node.hostname.get",
				Target:    "_all",
			})

			report, err := plan.Run(context.Background())

			s.Require().NoError(err)
			s.Require().Len(report.Tasks, 1)
			s.Len(
				report.Tasks[0].HostResults,
				tt.wantHostResults,
			)
			s.Equal(
				tt.wantHostname,
				report.Tasks[0].HostResults[0].Hostname,
			)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestExecuteOpCommandNonZeroExit() {
	tests := []struct {
		name      string
		operation string
		jobResult map[string]any
		wantErr   string
	}{
		{
			name:      "command exec with non-zero exit code fails",
			operation: "command.exec.execute",
			jobResult: map[string]any{
				"exit_code": float64(1),
				"stdout":    "",
				"stderr":    "command not found",
			},
			wantErr: "command exited with code 1",
		},
		{
			name:      "command shell with non-zero exit code fails",
			operation: "command.shell.execute",
			jobResult: map[string]any{
				"exit_code": float64(127),
				"stdout":    "",
				"stderr":    "not found",
			},
			wantErr: "command exited with code 127",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			origInterval := DefaultPollInterval
			DefaultPollInterval = 10 * time.Millisecond

			defer func() {
				DefaultPollInterval = origInterval
			}()

			srv := jobTestServer(tt.jobResult)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(Continue))

			plan.Task("cmd-op", &Op{
				Operation: tt.operation,
				Target:    "_any",
				Params:    map[string]any{"command": "false"},
			})

			report, err := plan.Run(context.Background())

			// With Continue strategy, run() doesn't return
			// the error, but the task result carries it.
			_ = err
			s.Require().Len(report.Tasks, 1)
			s.Equal(StatusFailed, report.Tasks[0].Status)
			s.Require().NotNil(report.Tasks[0].Error)
			s.Contains(
				report.Tasks[0].Error.Error(),
				tt.wantErr,
			)
		})
	}
}
