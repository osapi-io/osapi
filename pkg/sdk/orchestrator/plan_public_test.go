package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	osapiclient "github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type PlanPublicTestSuite struct {
	suite.Suite
}

func TestPlanPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PlanPublicTestSuite))
}

// pollResponse describes one GET /job/{id} response from opServer.
type pollResponse struct {
	status string // "pending", "completed", "failed", or "" (omit field)
	result any    // nil, map[string]any, string
	err    string // error message for failed jobs
	code   int    // HTTP status code (0 = 200)
}

// opServer creates an httptest.Server that handles job create + poll.
// createCode is the HTTP status for POST /job.
// pollResponses is a sequence of responses returned on successive GET
// /job/{id} calls.
func opServer(
	s *PlanPublicTestSuite,
	createCode int,
	createErrMsg string,
	pollResponses []pollResponse,
) *httptest.Server {
	s.T().Helper()

	var pollIdx atomic.Int32
	jobID := "00000000-0000-0000-0000-000000000001"

	return httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(createCode)
			if createCode == http.StatusCreated {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"job_id": jobID,
					"status": "pending",
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": createErrMsg,
				})
			}
		case r.Method == "GET" && r.URL.Path == "/job/"+jobID:
			idx := int(pollIdx.Add(1)) - 1
			if idx >= len(pollResponses) {
				idx = len(pollResponses) - 1
			}

			pr := pollResponses[idx]

			code := pr.code
			if code == 0 {
				code = http.StatusOK
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)

			resp := map[string]any{"id": jobID}
			if pr.status != "" {
				resp["status"] = pr.status
			}
			if pr.result != nil {
				resp["result"] = pr.result
			}
			if pr.err != "" {
				resp["error"] = pr.err
			}

			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// withShortPoll sets DefaultPollInterval to 10ms for the test duration.
func withShortPoll() func() {
	orig := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond

	return func() { orchestrator.DefaultPollInterval = orig }
}

// taskFunc creates a TaskFn with the given changed value and optional
// side effect.
func taskFunc(
	changed bool,
	sideEffect func(),
) orchestrator.TaskFn {
	return func(
		_ context.Context,
		_ *osapiclient.Client,
	) (*orchestrator.Result, error) {
		if sideEffect != nil {
			sideEffect()
		}

		return &orchestrator.Result{Changed: changed}, nil
	}
}

// failFunc creates a TaskFn that always returns the given error.
func failFunc(msg string) orchestrator.TaskFn {
	return func(
		_ context.Context,
		_ *osapiclient.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("%s", msg)
	}
}

// statusMap builds a name→status map from a report for easy assertions.
func statusMap(report *orchestrator.Report) map[string]orchestrator.Status {
	m := make(map[string]orchestrator.Status, len(report.Tasks))
	for _, r := range report.Tasks {
		m[r.Name] = r.Status
	}

	return m
}

// filterPrefix returns only strings that start with prefix.
func filterPrefix(
	ss []string,
	prefix string,
) []string {
	var out []string
	for _, s := range ss {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			out = append(out, s)
		}
	}

	return out
}

func (s *PlanPublicTestSuite) TestRun() {
	s.Run("linear chain executes in order", func() {
		var order []string
		plan := orchestrator.NewPlan(nil)

		mk := func(name string, changed bool) *orchestrator.Task {
			n := name

			return plan.TaskFunc(n, taskFunc(changed, func() {
				order = append(order, n)
			}))
		}

		a := mk("a", true)
		b := mk("b", true)
		c := mk("c", false)
		b.DependsOn(a)
		c.DependsOn(b)

		report, err := plan.Run(context.Background())
		s.Require().NoError(err)
		s.Equal([]string{"a", "b", "c"}, order)
		s.Len(report.Tasks, 3)
		s.Contains(report.Summary(), "2 changed")
		s.Contains(report.Summary(), "1 unchanged")
	})

	s.Run("parallel tasks run concurrently", func() {
		var concurrentMax atomic.Int32
		var concurrent atomic.Int32

		plan := orchestrator.NewPlan(nil)

		for _, name := range []string{"a", "b", "c"} {
			plan.TaskFunc(name, func(
				_ context.Context,
				_ *osapiclient.Client,
			) (*orchestrator.Result, error) {
				cur := concurrent.Add(1)
				for {
					prev := concurrentMax.Load()
					if cur > prev {
						if concurrentMax.CompareAndSwap(prev, cur) {
							break
						}
					} else {
						break
					}
				}
				concurrent.Add(-1)

				return &orchestrator.Result{Changed: false}, nil
			})
		}

		report, err := plan.Run(context.Background())
		s.Require().NoError(err)
		s.Len(report.Tasks, 3)
		s.GreaterOrEqual(int(concurrentMax.Load()), 1)
	})

	s.Run("cycle detection returns error", func() {
		plan := orchestrator.NewPlan(nil)
		a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
		b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
		a.DependsOn(b)
		b.DependsOn(a)

		_, err := plan.Run(context.Background())
		s.Error(err)
		s.Contains(err.Error(), "cycle")
	})
}

func (s *PlanPublicTestSuite) TestRunOnlyIfChanged() {
	tests := []struct {
		name         string
		depChanged   bool
		validateFunc func(report *orchestrator.Report, ran bool)
	}{
		{
			name:       "skips when no dependency changed",
			depChanged: false,
			validateFunc: func(report *orchestrator.Report, ran bool) {
				s.False(ran)
				s.Contains(report.Summary(), "skipped")
			},
		},
		{
			name:       "runs when dependency changed",
			depChanged: true,
			validateFunc: func(report *orchestrator.Report, ran bool) {
				s.True(ran)
				s.Equal(orchestrator.StatusUnchanged, report.Tasks[1].Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			ran := false

			dep := plan.TaskFunc("dep", taskFunc(tt.depChanged, nil))
			conditional := plan.TaskFunc("conditional", taskFunc(false, func() {
				ran = true
			}))
			conditional.DependsOn(dep).OnlyIfChanged()

			report, err := plan.Run(context.Background())
			s.Require().NoError(err)
			tt.validateFunc(report, ran)
		})
	}
}

func (s *PlanPublicTestSuite) TestRunGuard() {
	tests := []struct {
		name         string
		guard        func(orchestrator.Results) bool
		validateFunc func(report *orchestrator.Report, ran bool)
	}{
		{
			name:  "skips when guard returns false",
			guard: func(_ orchestrator.Results) bool { return false },
			validateFunc: func(report *orchestrator.Report, ran bool) {
				s.False(ran)
				s.Equal(orchestrator.StatusSkipped, report.Tasks[1].Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			ran := false

			a := plan.TaskFunc("a", taskFunc(false, nil))
			b := plan.TaskFunc("b", taskFunc(true, func() {
				ran = true
			}))
			b.DependsOn(a)
			b.When(tt.guard)

			report, err := plan.Run(context.Background())
			s.Require().NoError(err)
			tt.validateFunc(report, ran)
		})
	}
}

func (s *PlanPublicTestSuite) TestRunGuardWithFailedDependency() {
	tests := []struct {
		name         string
		guard        func(orchestrator.Results) bool
		expectRan    bool
		expectStatus orchestrator.Status
	}{
		{
			name: "guard runs and returns true when dependency failed",
			guard: func(r orchestrator.Results) bool {
				res := r.Get("fail")

				return res != nil && res.Status == orchestrator.StatusFailed
			},
			expectRan:    true,
			expectStatus: orchestrator.StatusChanged,
		},
		{
			name: "guard runs and returns false when dependency failed",
			guard: func(r orchestrator.Results) bool {
				res := r.Get("fail")

				return res != nil && res.Status == orchestrator.StatusChanged
			},
			expectRan:    false,
			expectStatus: orchestrator.StatusSkipped,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(
				nil,
				orchestrator.OnError(orchestrator.Continue),
			)
			ran := false

			fail := plan.TaskFunc("fail", failFunc("boom"))
			alert := plan.TaskFunc("alert", taskFunc(true, func() {
				ran = true
			}))
			alert.DependsOn(fail)
			alert.When(tt.guard)

			report, err := plan.Run(context.Background())
			s.Require().NoError(err)
			s.Equal(tt.expectRan, ran)

			sm := statusMap(report)
			s.Equal(tt.expectStatus, sm["alert"])
		})
	}
}

func (s *PlanPublicTestSuite) TestRunErrorStrategy() {
	s.Run("stop all on error", func() {
		plan := orchestrator.NewPlan(nil)
		didRun := false

		fail := plan.TaskFunc("fail", failFunc("boom"))
		next := plan.TaskFunc("next", taskFunc(false, func() {
			didRun = true
		}))
		next.DependsOn(fail)

		_, err := plan.Run(context.Background())
		s.Error(err)
		s.False(didRun)
	})

	s.Run("continue on error runs independent tasks", func() {
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Continue),
		)
		didRun := false

		plan.TaskFunc("fail", failFunc("boom"))
		plan.TaskFunc("independent", taskFunc(true, func() {
			didRun = true
		}))

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.True(didRun)
		s.Len(report.Tasks, 2)
	})

	s.Run("continue skips dependents of failed task", func() {
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Continue),
		)

		a := plan.TaskFunc("a", failFunc("a failed"))
		plan.TaskFunc("b", taskFunc(true, nil)).DependsOn(a)
		plan.TaskFunc("c", taskFunc(true, nil))

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.Len(report.Tasks, 3)
		m := statusMap(report)
		s.Equal(orchestrator.StatusFailed, m["a"])
		s.Equal(orchestrator.StatusSkipped, m["b"])
		s.Equal(orchestrator.StatusChanged, m["c"])
	})

	s.Run("continue transitive skip", func() {
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Continue),
		)

		a := plan.TaskFunc("a", failFunc("a failed"))
		b := plan.TaskFunc("b", taskFunc(true, nil))
		b.DependsOn(a)
		c := plan.TaskFunc("c", taskFunc(true, nil))
		c.DependsOn(b)

		report, err := plan.Run(context.Background())
		s.NoError(err)
		m := statusMap(report)
		s.Equal(orchestrator.StatusFailed, m["a"])
		s.Equal(orchestrator.StatusSkipped, m["b"])
		s.Equal(orchestrator.StatusSkipped, m["c"])
	})

	s.Run("per-task continue override", func() {
		plan := orchestrator.NewPlan(nil) // default StopAll

		a := plan.TaskFunc("a", failFunc("a failed"))
		a.OnError(orchestrator.Continue)
		plan.TaskFunc("b", taskFunc(true, nil))

		report, err := plan.Run(context.Background())
		s.NoError(err)
		m := statusMap(report)
		s.Equal(orchestrator.StatusFailed, m["a"])
		s.Equal(orchestrator.StatusChanged, m["b"])
	})

	s.Run("retry succeeds after transient failure", func() {
		attempts := 0
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Retry(2)),
		)

		plan.TaskFunc("flaky", func(
			_ context.Context,
			_ *osapiclient.Client,
		) (*orchestrator.Result, error) {
			attempts++
			if attempts < 3 {
				return nil, fmt.Errorf("attempt %d failed", attempts)
			}

			return &orchestrator.Result{Changed: true}, nil
		})

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.Equal(3, attempts)
		s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
	})

	s.Run("retry exhausted returns error", func() {
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Retry(1)),
		)

		plan.TaskFunc("always-fail", failFunc("permanent failure"))

		report, err := plan.Run(context.Background())
		s.Error(err)
		s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
	})

	s.Run("per-task retry override", func() {
		attempts := 0
		plan := orchestrator.NewPlan(nil) // default StopAll

		plan.TaskFunc("flaky", func(
			_ context.Context,
			_ *osapiclient.Client,
		) (*orchestrator.Result, error) {
			attempts++
			if attempts < 2 {
				return nil, fmt.Errorf("attempt %d failed", attempts)
			}

			return &orchestrator.Result{Changed: true}, nil
		}).OnError(orchestrator.Retry(1))

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.Equal(2, attempts)
		s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
	})
}

func (s *PlanPublicTestSuite) TestRunOpTask() {
	tests := []struct {
		name          string
		createCode    int
		createErrMsg  string
		pollResponses []pollResponse
		op            *orchestrator.Op
		noServer      bool
		validateFunc  func(report *orchestrator.Report, err error)
	}{
		{
			name:       "completed with changed true in result data",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "pending"},
				{status: "completed", result: map[string]any{
					"changed": true,
					"success": true,
					"message": "DNS updated",
				}},
			},
			op: &orchestrator.Op{
				Operation: "network.dns.update",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Len(report.Tasks, 1)
				s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
				s.True(report.Tasks[0].Changed)
			},
		},
		{
			name:       "completed with changed false in result data",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "pending"},
				{status: "completed", result: map[string]any{
					"hostname": "web-01",
					"changed":  false,
				}},
			},
			op: &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Len(report.Tasks, 1)
				s.Equal(orchestrator.StatusUnchanged, report.Tasks[0].Status)
				s.False(report.Tasks[0].Changed)
			},
		},
		{
			name:       "completed with no changed field defaults to false",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "completed", result: map[string]any{"hostname": "web-01"}},
			},
			op: &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Equal(orchestrator.StatusUnchanged, report.Tasks[0].Status)
				s.False(report.Tasks[0].Changed)
			},
		},
		{
			name:       "completed with no result data defaults to unchanged",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "completed"},
			},
			op: &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Equal(orchestrator.StatusUnchanged, report.Tasks[0].Status)
				s.False(report.Tasks[0].Changed)
			},
		},
		{
			name:       "completed with non-map result defaults to unchanged",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "completed", result: "just a string"},
			},
			op: &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Equal(orchestrator.StatusUnchanged, report.Tasks[0].Status)
				s.False(report.Tasks[0].Changed)
			},
		},
		{
			name:       "polls past nil status",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: ""},
				{status: "completed", result: map[string]any{"changed": true}},
			},
			op: &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Require().NoError(err)
				s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
			},
		},
		{
			name:     "requires client",
			noServer: true,
			op: &orchestrator.Op{
				Operation: "command.exec",
				Target:    "_any",
				Params:    map[string]any{"command": "uptime"},
			},
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.NotNil(report)
				s.Contains(report.Tasks[0].Error.Error(), "requires an OSAPI client")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.noServer {
				plan := orchestrator.NewPlan(nil)
				plan.Task("op-task", tt.op)

				report, err := plan.Run(context.Background())
				tt.validateFunc(report, err)

				return
			}

			restore := withShortPoll()
			defer restore()

			srv := opServer(s, tt.createCode, tt.createErrMsg, tt.pollResponses)
			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")

			plan := orchestrator.NewPlan(client)
			plan.Task("op-task", tt.op)

			report, runErr := plan.Run(context.Background())
			tt.validateFunc(report, runErr)
		})
	}
}

func (s *PlanPublicTestSuite) TestRunOpTaskParams() {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-00000000000a",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-00000000000a",
				"status": "completed",
			})
		}
	}))
	defer srv.Close()

	restore := withShortPoll()
	defer restore()

	client := osapiclient.New(srv.URL, "test-token")

	plan := orchestrator.NewPlan(client)
	plan.Task("exec", &orchestrator.Op{
		Operation: "command.exec.execute",
		Target:    "_any",
		Params: map[string]any{
			"command": "uptime",
			"args":    []string{"-s"},
		},
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal(orchestrator.StatusUnchanged, report.Tasks[0].Status)

	op, ok := receivedBody["operation"].(map[string]any)
	s.Require().True(ok)
	s.Equal("command.exec.execute", op["type"])

	data, ok := op["data"].(map[string]any)
	s.Require().True(ok)
	s.Equal("uptime", data["command"])
}

func (s *PlanPublicTestSuite) TestRunOpTaskErrors() {
	tests := []struct {
		name          string
		createCode    int
		createErrMsg  string
		pollResponses []pollResponse
		useServer     bool
		networkError  string // "create" or "poll"
		useContext    bool
		pollInterval  time.Duration // overrides withShortPoll if set
		validateFunc  func(report *orchestrator.Report, err error)
	}{
		{
			name:         "create returns server error",
			createCode:   http.StatusInternalServerError,
			createErrMsg: "internal server error",
			useServer:    true,
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.NotNil(report)
				s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
				s.Contains(err.Error(), "internal server error")
			},
		},
		{
			name:         "create returns validation error",
			createCode:   http.StatusBadRequest,
			createErrMsg: "validation failed on 'target_hostname' for 'valid_target'",
			useServer:    true,
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.NotNil(report)
				s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
				s.Contains(err.Error(), "valid_target")
			},
		},
		{
			name:       "poll returns server error",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{code: http.StatusInternalServerError},
			},
			useServer: true,
			validateFunc: func(_ *orchestrator.Report, err error) {
				s.Error(err)
				s.Contains(err.Error(), "poll job")
			},
		},
		{
			name:       "job failed with error message",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "failed", err: "disk full"},
			},
			useServer: true,
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.Contains(err.Error(), "disk full")
				s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
			},
		},
		{
			name:       "job failed without error message",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "failed"},
			},
			useServer: true,
			validateFunc: func(_ *orchestrator.Report, err error) {
				s.Error(err)
				s.Contains(err.Error(), "job failed")
			},
		},
		{
			name:       "context canceled during poll",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "pending"},
			},
			useServer:  true,
			useContext: true,
			validateFunc: func(_ *orchestrator.Report, err error) {
				s.Error(err)
				s.ErrorIs(err, context.DeadlineExceeded)
			},
		},
		{
			name:         "create network error",
			networkError: "create",
			validateFunc: func(_ *orchestrator.Report, err error) {
				s.Error(err)
				s.Contains(err.Error(), "create job")
			},
		},
		{
			name:         "poll network error",
			networkError: "poll",
			validateFunc: func(_ *orchestrator.Report, err error) {
				s.Error(err)
				s.Contains(err.Error(), "poll job")
			},
		},
		{
			name:       "create returns unexpected status",
			createCode: http.StatusOK,
			useServer:  true,
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.NotNil(report)
				s.Contains(err.Error(), "nil response body")
			},
		},
		{
			name:       "context canceled before first poll tick",
			createCode: http.StatusCreated,
			pollResponses: []pollResponse{
				{status: "pending"},
			},
			useServer:    true,
			useContext:   true,
			pollInterval: 5 * time.Second,
			validateFunc: func(report *orchestrator.Report, err error) {
				s.Error(err)
				s.NotNil(report)
				s.ErrorIs(err, context.DeadlineExceeded)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var restore func()
			if tt.pollInterval > 0 {
				orig := orchestrator.DefaultPollInterval
				orchestrator.DefaultPollInterval = tt.pollInterval
				restore = func() { orchestrator.DefaultPollInterval = orig }
			} else {
				restore = withShortPoll()
			}
			defer restore()

			var srv *httptest.Server

			switch {
			case tt.networkError == "create":
				srv = httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					_ *http.Request,
				) {
					hj, ok := w.(http.Hijacker)
					if ok {
						conn, _, _ := hj.Hijack()
						_ = conn.Close()
					}
				}))
			case tt.networkError == "poll":
				srv = httptest.NewServer(http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					if r.Method == "POST" && r.URL.Path == "/job" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(map[string]any{
							"job_id": "00000000-0000-0000-0000-000000000009",
							"status": "pending",
						})

						return
					}

					hj, ok := w.(http.Hijacker)
					if ok {
						conn, _, _ := hj.Hijack()
						_ = conn.Close()
					}
				}))
			case tt.useServer:
				srv = opServer(s, tt.createCode, tt.createErrMsg, tt.pollResponses)
			}

			defer srv.Close()

			client := osapiclient.New(srv.URL, "test-token")

			plan := orchestrator.NewPlan(client)
			plan.Task("op-task", &orchestrator.Op{
				Operation: "node.hostname.get",
				Target:    "_any",
			})

			ctx := context.Background()
			if tt.useContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			report, runErr := plan.Run(ctx)
			tt.validateFunc(report, runErr)
		})
	}
}

func (s *PlanPublicTestSuite) TestRunHooks() {
	s.Run("all hooks called in order", func() {
		var events []string

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

		a := plan.TaskFunc("a", taskFunc(true, nil))
		plan.TaskFunc("b", taskFunc(false, nil)).DependsOn(a)

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.NotNil(report)
		s.Equal([]string{
			"before-plan",
			"before-level-0",
			"before-a",
			"after-a",
			"after-level-0",
			"before-level-1",
			"before-b",
			"after-b",
			"after-level-1",
			"after-plan",
		}, events)
	})

	s.Run("retry hook called with correct args", func() {
		var events []string
		attempts := 0

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.WithHooks(hooks),
			orchestrator.OnError(orchestrator.Retry(2)),
		)

		plan.TaskFunc("flaky", func(
			_ context.Context,
			_ *osapiclient.Client,
		) (*orchestrator.Result, error) {
			attempts++
			if attempts < 3 {
				return nil, fmt.Errorf("fail-%d", attempts)
			}

			return &orchestrator.Result{Changed: true}, nil
		})

		_, err := plan.Run(context.Background())
		s.NoError(err)

		retries := filterPrefix(events, "retry-")
		s.Len(retries, 2)
		s.Contains(retries[0], "retry-flaky-1-fail-1")
		s.Contains(retries[1], "retry-flaky-2-fail-2")
	})

	s.Run("skip hook for dependency failure", func() {
		var events []string

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.WithHooks(hooks),
			orchestrator.OnError(orchestrator.Continue),
		)

		a := plan.TaskFunc("a", failFunc("a failed"))
		plan.TaskFunc("b", taskFunc(true, nil)).DependsOn(a)

		_, err := plan.Run(context.Background())
		s.NoError(err)

		skips := filterPrefix(events, "skip-")
		s.Len(skips, 1)
		s.Contains(skips[0], "skip-b-dependency failed")
	})

	s.Run("skip hook for guard", func() {
		var events []string

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

		a := plan.TaskFunc("a", taskFunc(false, nil))
		b := plan.TaskFunc("b", taskFunc(true, nil))
		b.DependsOn(a)
		b.When(func(_ orchestrator.Results) bool { return false })

		_, err := plan.Run(context.Background())
		s.NoError(err)

		skips := filterPrefix(events, "skip-")
		s.Len(skips, 1)
		s.Contains(skips[0], "guard returned false")
	})

	s.Run("skip hook for guard with custom reason", func() {
		var events []string

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

		a := plan.TaskFunc("a", taskFunc(false, nil))
		b := plan.TaskFunc("b", taskFunc(true, nil))
		b.DependsOn(a)
		b.WhenWithReason(
			func(_ orchestrator.Results) bool { return false },
			"host is unreachable",
		)

		_, err := plan.Run(context.Background())
		s.NoError(err)

		skips := filterPrefix(events, "skip-")
		s.Len(skips, 1)
		s.Contains(skips[0], "host is unreachable")
	})

	s.Run("skip hook for only-if-changed", func() {
		var events []string

		hooks := allHooks(&events)
		plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

		a := plan.TaskFunc("a", taskFunc(false, nil))
		b := plan.TaskFunc("b", taskFunc(true, nil))
		b.DependsOn(a)
		b.OnlyIfChanged()

		_, err := plan.Run(context.Background())
		s.NoError(err)

		skips := filterPrefix(events, "skip-")
		s.Len(skips, 1)
		s.Contains(skips[0], "no dependencies changed")
	})

	s.Run("retry without hooks configured", func() {
		attempts := 0
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Retry(1)),
		)

		plan.TaskFunc("flaky", func(
			_ context.Context,
			_ *osapiclient.Client,
		) (*orchestrator.Result, error) {
			attempts++
			if attempts < 2 {
				return nil, fmt.Errorf("attempt %d", attempts)
			}

			return &orchestrator.Result{Changed: true}, nil
		})

		report, err := plan.Run(context.Background())
		s.NoError(err)
		s.Equal(2, attempts)
		s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
	})

	s.Run("skip without hooks configured", func() {
		plan := orchestrator.NewPlan(
			nil,
			orchestrator.OnError(orchestrator.Continue),
		)

		a := plan.TaskFunc("a", failFunc("a failed"))
		plan.TaskFunc("b", taskFunc(true, nil)).DependsOn(a)

		report, err := plan.Run(context.Background())
		s.NoError(err)
		m := statusMap(report)
		s.Equal(orchestrator.StatusSkipped, m["b"])
	})
}

func (s *PlanPublicTestSuite) TestClient() {
	client := osapiclient.New("http://localhost", "token")

	plan := orchestrator.NewPlan(client)
	s.Equal(client, plan.Client())

	nilPlan := orchestrator.NewPlan(nil)
	s.Nil(nilPlan.Client())
}

func (s *PlanPublicTestSuite) TestConfig() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	cfg := plan.Config()
	s.Equal("continue", cfg.OnErrorStrategy.String())
}

func (s *PlanPublicTestSuite) TestTasks() {
	plan := orchestrator.NewPlan(nil)
	s.Empty(plan.Tasks())

	plan.TaskFunc("a", taskFunc(false, nil))
	plan.TaskFunc("b", taskFunc(false, nil))
	s.Len(plan.Tasks(), 2)
}

func (s *PlanPublicTestSuite) TestValidate() {
	tests := []struct {
		name         string
		setup        func(plan *orchestrator.Plan)
		validateFunc func(err error)
	}{
		{
			name: "duplicate task name returns error",
			setup: func(plan *orchestrator.Plan) {
				plan.TaskFunc("dup", taskFunc(false, nil))
				plan.TaskFunc("dup", taskFunc(false, nil))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "duplicate task name")
			},
		},
		{
			name: "valid plan returns nil",
			setup: func(plan *orchestrator.Plan) {
				plan.TaskFunc("a", taskFunc(false, nil))
				plan.TaskFunc("b", taskFunc(false, nil))
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			tt.setup(plan)
			tt.validateFunc(plan.Validate())
		})
	}
}

func (s *PlanPublicTestSuite) TestLevels() {
	tests := []struct {
		name         string
		setup        func(plan *orchestrator.Plan)
		validateFunc func(levels [][]*orchestrator.Task, err error)
	}{
		{
			name: "returns levels for valid plan",
			setup: func(plan *orchestrator.Plan) {
				a := plan.TaskFunc("a", taskFunc(false, nil))
				plan.TaskFunc("b", taskFunc(false, nil)).DependsOn(a)
			},
			validateFunc: func(levels [][]*orchestrator.Task, err error) {
				s.NoError(err)
				s.Len(levels, 2)
			},
		},
		{
			name: "returns error for invalid plan",
			setup: func(plan *orchestrator.Plan) {
				a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
				b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
				a.DependsOn(b)
				b.DependsOn(a)
			},
			validateFunc: func(levels [][]*orchestrator.Task, err error) {
				s.Error(err)
				s.Nil(levels)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			tt.setup(plan)
			levels, err := plan.Levels()
			tt.validateFunc(levels, err)
		})
	}
}

func (s *PlanPublicTestSuite) TestExplain() {
	tests := []struct {
		name     string
		setup    func(plan *orchestrator.Plan)
		contains []string
	}{
		{
			name: "valid plan with dependencies and guards",
			setup: func(plan *orchestrator.Plan) {
				a := plan.TaskFunc("a", taskFunc(false, nil))
				b := plan.TaskFunc("b", taskFunc(false, nil))
				b.DependsOn(a)
				b.OnlyIfChanged()
			},
			contains: []string{
				"Plan: 2 tasks, 2 levels",
				"Level 0:",
				"a [fn]",
				"Level 1:",
				"b [fn]",
				"only-if-changed",
			},
		},
		{
			name: "invalid plan returns error string",
			setup: func(plan *orchestrator.Plan) {
				a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
				b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
				a.DependsOn(b)
				b.DependsOn(a)
			},
			contains: []string{"invalid plan:", "cycle"},
		},
		{
			name: "parallel tasks shown as parallel",
			setup: func(plan *orchestrator.Plan) {
				plan.TaskFunc("a", taskFunc(false, nil))
				plan.TaskFunc("b", taskFunc(false, nil))
			},
			contains: []string{
				"Plan: 2 tasks, 1 levels",
				"Level 0 (parallel):",
			},
		},
		{
			name: "op task shown as op",
			setup: func(plan *orchestrator.Plan) {
				plan.Task("install", &orchestrator.Op{
					Operation: "node.hostname.get",
				})
			},
			contains: []string{"install [op]"},
		},
		{
			name: "guard shown in flags",
			setup: func(plan *orchestrator.Plan) {
				a := plan.TaskFunc("a", taskFunc(false, nil))
				b := plan.TaskFunc("b", taskFunc(false, nil))
				b.DependsOn(a)
				b.When(func(_ orchestrator.Results) bool { return true })
			},
			contains: []string{"when"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(nil)
			tt.setup(plan)
			output := plan.Explain()
			for _, c := range tt.contains {
				s.Contains(output, c)
			}
		})
	}
}

// allHooks returns a Hooks struct that appends all events to the given
// slice, covering every hook type.
func allHooks(events *[]string) orchestrator.Hooks {
	return orchestrator.Hooks{
		BeforePlan: func(_ orchestrator.PlanSummary) {
			*events = append(*events, "before-plan")
		},
		AfterPlan: func(_ *orchestrator.Report) {
			*events = append(*events, "after-plan")
		},
		BeforeLevel: func(level int, _ []*orchestrator.Task, _ bool) {
			*events = append(*events, fmt.Sprintf("before-level-%d", level))
		},
		AfterLevel: func(level int, _ []orchestrator.TaskResult) {
			*events = append(*events, fmt.Sprintf("after-level-%d", level))
		},
		BeforeTask: func(task *orchestrator.Task) {
			*events = append(*events, "before-"+task.Name())
		},
		AfterTask: func(_ *orchestrator.Task, r orchestrator.TaskResult) {
			*events = append(*events, "after-"+r.Name)
		},
		OnRetry: func(
			task *orchestrator.Task,
			attempt int,
			err error,
		) {
			*events = append(
				*events,
				fmt.Sprintf("retry-%s-%d-%s", task.Name(), attempt, err),
			)
		},
		OnSkip: func(task *orchestrator.Task, reason string) {
			*events = append(
				*events,
				fmt.Sprintf("skip-%s-%s", task.Name(), reason),
			)
		},
	}
}
