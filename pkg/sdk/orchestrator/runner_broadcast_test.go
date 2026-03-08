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
			got := extractHostResults(tt.data, nil)
			s.Equal(tt.want, got)
		})
	}

	s.Run("enriches host result with agent duration", func() {
		data := map[string]any{
			"results": []any{
				map[string]any{
					"hostname": "host-1",
					"changed":  true,
				},
			},
		}
		durations := map[string]time.Duration{
			"host-1": 3 * time.Second,
		}

		got := extractHostResults(data, durations)
		s.Require().Len(got, 1)
		s.Equal(3*time.Second, got[0].JobDuration)
	})
}

func (s *RunnerBroadcastTestSuite) TestHostResultsFromResponses() {
	changed := true

	tests := []struct {
		name      string
		responses map[string]osapiclient.AgentJobResponse
		durations map[string]time.Duration
		wantLen   int
		validate  func(hrs []HostResult)
	}{
		{
			name: "populates Changed Error Data and JobDuration",
			responses: map[string]osapiclient.AgentJobResponse{
				"web-01": {
					Hostname: "web-01",
					Changed:  &changed,
					Error:    "timeout",
					Data:     map[string]any{"stdout": "hello"},
				},
			},
			durations: map[string]time.Duration{
				"web-01": 5 * time.Second,
			},
			wantLen: 1,
			validate: func(hrs []HostResult) {
				s.True(hrs[0].Changed)
				s.Equal("timeout", hrs[0].Error)
				s.Equal("hello", hrs[0].Data["stdout"])
				s.Equal(5*time.Second, hrs[0].JobDuration)
			},
		},
		{
			name: "nil Changed and nil Data and no duration",
			responses: map[string]osapiclient.AgentJobResponse{
				"web-02": {
					Hostname: "web-02",
				},
			},
			durations: nil,
			wantLen:   1,
			validate: func(hrs []HostResult) {
				s.False(hrs[0].Changed)
				s.Empty(hrs[0].Error)
				s.Nil(hrs[0].Data)
				s.Zero(hrs[0].JobDuration)
			},
		},
		{
			name: "Data that is not map is ignored",
			responses: map[string]osapiclient.AgentJobResponse{
				"web-03": {
					Hostname: "web-03",
					Data:     "not-a-map",
				},
			},
			wantLen: 1,
			validate: func(hrs []HostResult) {
				s.Nil(hrs[0].Data)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := hostResultsFromResponses(tt.responses, tt.durations)
			s.Len(got, tt.wantLen)
			if tt.validate != nil {
				tt.validate(got)
			}
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestParseAgentDurations() {
	tests := []struct {
		name        string
		states      map[string]osapiclient.AgentState
		wantLongest time.Duration
		wantLen     int
		validate    func(perHost map[string]time.Duration)
	}{
		{
			name: "parses valid durations and finds longest",
			states: map[string]osapiclient.AgentState{
				"web-01": {Duration: "2s"},
				"web-02": {Duration: "5s"},
			},
			wantLongest: 5 * time.Second,
			wantLen:     2,
			validate: func(perHost map[string]time.Duration) {
				s.Equal(2*time.Second, perHost["web-01"])
				s.Equal(5*time.Second, perHost["web-02"])
			},
		},
		{
			name: "skips empty duration",
			states: map[string]osapiclient.AgentState{
				"web-01": {Duration: ""},
			},
			wantLongest: 0,
			wantLen:     0,
		},
		{
			name: "skips invalid duration",
			states: map[string]osapiclient.AgentState{
				"web-01": {Duration: "not-a-duration"},
			},
			wantLongest: 0,
			wantLen:     0,
		},
		{
			name:        "empty states returns zero",
			states:      map[string]osapiclient.AgentState{},
			wantLongest: 0,
			wantLen:     0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			longest, perHost := parseAgentDurations(tt.states)
			s.Equal(tt.wantLongest, longest)
			s.Len(perHost, tt.wantLen)
			if tt.validate != nil {
				tt.validate(perHost)
			}
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

// jobTestServerWithChanged creates an httptest server where the
// job-level changed field is explicitly set.
func jobTestServerWithChanged(
	jobResult map[string]any,
	changed bool,
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
				"id":      jobID,
				"status":  "completed",
				"result":  jobResult,
				"changed": changed,
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// jobTestServerWithResponses creates an httptest server that returns
// per-agent Responses in the job detail (simulating the API with
// responses exposed for all targets).
func jobTestServerWithResponses(
	jobResult map[string]any,
	responses map[string]any,
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
				"id":        jobID,
				"status":    "completed",
				"result":    jobResult,
				"responses": responses,
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func (s *RunnerBroadcastTestSuite) TestExecuteOpUnicastClearsHostResults() {
	tests := []struct {
		name      string
		target    string
		responses map[string]any
	}{
		{
			name:   "any target clears host results from responses",
			target: "_any",
			responses: map[string]any{
				"nerd": map[string]any{
					"hostname": "nerd",
					"status":   "completed",
					"data":     map[string]any{"hostname": "nerd"},
				},
			},
		},
		{
			name:   "specific host target clears host results",
			target: "web-01",
			responses: map[string]any{
				"web-01": map[string]any{
					"hostname": "web-01",
					"status":   "completed",
					"data":     map[string]any{"hostname": "web-01"},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			origInterval := DefaultPollInterval
			DefaultPollInterval = 10 * time.Millisecond

			defer func() {
				DefaultPollInterval = origInterval
			}()

			srv := jobTestServerWithResponses(
				map[string]any{"hostname": "nerd"},
				tt.responses,
			)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(StopAll))

			plan.Task("get-hostname", &Op{
				Operation: "node.hostname.get",
				Target:    tt.target,
			})

			report, err := plan.Run(context.Background())

			s.Require().NoError(err)
			s.Require().Len(report.Tasks, 1)
			s.Nil(
				report.Tasks[0].HostResults,
				"non-broadcast target should not have host results",
			)
		})
	}
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

func (s *RunnerBroadcastTestSuite) TestExecuteOpBroadcastPartialFailure() {
	tests := []struct {
		name            string
		wantStatus      Status
		wantChanged     bool
		wantHostResults int
		wantHostError   string
	}{
		{
			name:            "broadcast with partial failure preserves Changed and HostResults",
			wantStatus:      StatusFailed,
			wantChanged:     true,
			wantHostResults: 2,
			wantHostError:   "command exited with code 1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			origInterval := DefaultPollInterval
			DefaultPollInterval = 10 * time.Millisecond

			defer func() {
				DefaultPollInterval = origInterval
			}()

			srv := jobTestServerWithChanged(
				map[string]any{
					"results": []any{
						map[string]any{
							"hostname":  "host-1",
							"changed":   true,
							"exit_code": float64(0),
							"stdout":    "deployed",
						},
						map[string]any{
							"hostname":  "host-2",
							"changed":   true,
							"exit_code": float64(1),
							"stderr":    "deploy failed",
						},
					},
				},
				true,
			)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(Continue))

			plan.Task("broadcast-cmd", &Op{
				Operation: "command.exec.execute",
				Target:    "_all",
				Params:    map[string]any{"command": "deploy.sh"},
			})

			report, err := plan.Run(context.Background())

			_ = err
			s.Require().Len(report.Tasks, 1)
			s.Equal(tt.wantStatus, report.Tasks[0].Status)
			s.Equal(tt.wantChanged, report.Tasks[0].Changed)
			s.Len(report.Tasks[0].HostResults, tt.wantHostResults)

			// Verify the failed host has the expected error.
			var foundError bool
			for _, hr := range report.Tasks[0].HostResults {
				if hr.Hostname == "host-2" {
					s.Contains(hr.Error, tt.wantHostError)
					foundError = true
				}
			}

			s.True(foundError, "should find host-2 with error")
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestCountExpectedAgents() {
	tests := []struct {
		name      string
		target    string
		agentResp string
		wantCount int
	}{
		{
			name:      "non-broadcast returns zero",
			target:    "web-01",
			agentResp: `{}`,
			wantCount: 0,
		},
		{
			name:   "all target counts non-cordoned agents",
			target: "_all",
			agentResp: `{
				"agents": [
					{"hostname":"web-01","status":"Ready"},
					{"hostname":"web-02","status":"Ready","state":"Cordoned"},
					{"hostname":"web-03","status":"Ready"}
				],
				"total": 3
			}`,
			wantCount: 2,
		},
		{
			name:   "all target excludes draining agents",
			target: "_all",
			agentResp: `{
				"agents": [
					{"hostname":"web-01","status":"Ready","state":"Draining"}
				],
				"total": 1
			}`,
			wantCount: 0,
		},
		{
			name:   "label selector counts matching agents",
			target: "group:web",
			agentResp: `{
				"agents": [
					{"hostname":"web-01","status":"Ready","labels":{"group":"web"}},
					{"hostname":"web-02","status":"Ready","labels":{"group":"db"}},
					{"hostname":"web-03","status":"Ready","labels":{"group":"web.prod"}}
				],
				"total": 3
			}`,
			wantCount: 2,
		},
		{
			name:   "label selector excludes cordoned",
			target: "group:web",
			agentResp: `{
				"agents": [
					{"hostname":"web-01","status":"Ready","state":"Cordoned","labels":{"group":"web"}}
				],
				"total": 1
			}`,
			wantCount: 0,
		},
		{
			name:   "label selector with no matching agents returns zero",
			target: "group:missing",
			agentResp: `{
				"agents": [
					{"hostname":"web-01","status":"Ready","labels":{"group":"web"}}
				],
				"total": 1
			}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			srv := httptest.NewServer(http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.URL.Path == "/agent" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.agentResp))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")
			plan := NewPlan(client, OnError(StopAll))
			r := newRunner(plan)

			got := r.countExpectedAgents(
				context.Background(),
				tt.target,
			)
			s.Equal(tt.wantCount, got)
		})
	}

	s.Run("API error returns zero", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(
			w http.ResponseWriter,
			_ *http.Request,
		) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		}))
		defer srv.Close()

		client := osapiclient.New(srv.URL, "test-token")
		plan := NewPlan(client, OnError(StopAll))
		r := newRunner(plan)

		got := r.countExpectedAgents(context.Background(), "_all")
		s.Equal(0, got)
	})
}

func (s *RunnerBroadcastTestSuite) TestPollJobBroadcastWaitsForAgents() {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			callCount++

			resp := map[string]any{
				"id":      "22222222-2222-2222-2222-222222222222",
				"status":  "completed",
				"result":  map[string]any{"hostname": "web-01"},
				"changed": true,
			}

			// First poll: only 1 response, but we expect 2.
			// Second poll: 2 responses → done.
			if callCount == 1 {
				resp["responses"] = map[string]any{
					"web-01": map[string]any{
						"hostname": "web-01",
						"status":   "completed",
					},
				}
			} else {
				resp["responses"] = map[string]any{
					"web-01": map[string]any{
						"hostname": "web-01",
						"status":   "completed",
					},
					"web-02": map[string]any{
						"hostname": "web-02",
						"status":   "completed",
					},
				}
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	origInterval := DefaultPollInterval
	DefaultPollInterval = 10 * time.Millisecond

	defer func() {
		DefaultPollInterval = origInterval
	}()

	client := osapiclient.New(srv.URL, "test-token")
	plan := NewPlan(client, OnError(StopAll))
	r := newRunner(plan)

	result, err := r.pollJob(
		context.Background(),
		"22222222-2222-2222-2222-222222222222",
		2,
	)

	s.Require().NoError(err)
	s.NotNil(result)
	s.True(result.Changed)
	s.GreaterOrEqual(callCount, 2, "should have polled at least twice")
	s.Len(result.HostResults, 2)
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
