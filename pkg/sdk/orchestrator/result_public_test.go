package orchestrator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type ResultPublicTestSuite struct {
	suite.Suite
}

func TestResultPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ResultPublicTestSuite))
}

func (s *ResultPublicTestSuite) TestReportSummary() {
	tests := []struct {
		name     string
		tasks    []orchestrator.TaskResult
		contains []string
	}{
		{
			name: "mixed results",
			tasks: []orchestrator.TaskResult{
				{
					Name:     "a",
					Status:   orchestrator.StatusChanged,
					Changed:  true,
					Duration: time.Second,
				},
				{
					Name:     "b",
					Status:   orchestrator.StatusUnchanged,
					Changed:  false,
					Duration: 2 * time.Second,
				},
				{Name: "c", Status: orchestrator.StatusSkipped, Changed: false, Duration: 0},
				{
					Name:     "d",
					Status:   orchestrator.StatusChanged,
					Changed:  true,
					Duration: 500 * time.Millisecond,
				},
			},
			contains: []string{"4 tasks", "2 changed", "1 unchanged", "1 skipped"},
		},
		{
			name: "all statuses including failed",
			tasks: []orchestrator.TaskResult{
				{Name: "a", Status: orchestrator.StatusChanged, Changed: true},
				{Name: "b", Status: orchestrator.StatusUnchanged},
				{Name: "c", Status: orchestrator.StatusSkipped},
				{Name: "d", Status: orchestrator.StatusFailed},
			},
			contains: []string{"4 tasks", "1 changed", "1 unchanged", "1 skipped", "1 failed"},
		},
		{
			name:     "empty report",
			tasks:    nil,
			contains: []string{"0 tasks"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			report := orchestrator.Report{Tasks: tt.tasks}
			summary := report.Summary()
			for _, c := range tt.contains {
				s.Contains(summary, c)
			}
		})
	}
}

func (s *ResultPublicTestSuite) TestResultStatusField() {
	tests := []struct {
		name       string
		result     *orchestrator.Result
		wantStatus orchestrator.Status
		wantChange bool
	}{
		{
			name: "changed result carries status",
			result: &orchestrator.Result{
				Changed: true,
				Data:    map[string]any{"hostname": "web-01"},
				Status:  orchestrator.StatusChanged,
			},
			wantStatus: orchestrator.StatusChanged,
			wantChange: true,
		},
		{
			name: "unchanged result carries status",
			result: &orchestrator.Result{
				Changed: false,
				Status:  orchestrator.StatusUnchanged,
			},
			wantStatus: orchestrator.StatusUnchanged,
			wantChange: false,
		},
		{
			name: "failed result carries status",
			result: &orchestrator.Result{
				Changed: false,
				Status:  orchestrator.StatusFailed,
			},
			wantStatus: orchestrator.StatusFailed,
			wantChange: false,
		},
		{
			name: "skipped result carries status",
			result: &orchestrator.Result{
				Changed: false,
				Status:  orchestrator.StatusSkipped,
			},
			wantStatus: orchestrator.StatusSkipped,
			wantChange: false,
		},
		{
			name:       "zero value has empty status",
			result:     &orchestrator.Result{},
			wantStatus: "",
			wantChange: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.wantStatus, tt.result.Status)
			s.Equal(tt.wantChange, tt.result.Changed)
		})
	}
}

func (s *ResultPublicTestSuite) TestResultHostResults() {
	tests := []struct {
		name       string
		result     *orchestrator.Result
		wantLen    int
		validateFn func(hrs []orchestrator.HostResult)
	}{
		{
			name: "result with multiple host results",
			result: &orchestrator.Result{
				Changed: true,
				Status:  orchestrator.StatusChanged,
				HostResults: []orchestrator.HostResult{
					{
						Hostname: "web-01",
						Changed:  true,
						Data:     map[string]any{"stdout": "ok"},
					},
					{
						Hostname: "web-02",
						Changed:  false,
						Error:    "connection timeout",
					},
				},
			},
			wantLen: 2,
			validateFn: func(hrs []orchestrator.HostResult) {
				s.Equal("web-01", hrs[0].Hostname)
				s.True(hrs[0].Changed)
				s.Equal("web-02", hrs[1].Hostname)
				s.Equal("connection timeout", hrs[1].Error)
			},
		},
		{
			name: "result with no host results",
			result: &orchestrator.Result{
				Changed: false,
				Status:  orchestrator.StatusUnchanged,
			},
			wantLen: 0,
		},
		{
			name: "host result with data map",
			result: &orchestrator.Result{
				Changed: true,
				Status:  orchestrator.StatusChanged,
				HostResults: []orchestrator.HostResult{
					{
						Hostname: "db-01",
						Changed:  true,
						Data: map[string]any{
							"stdout":    "migrated",
							"exit_code": float64(0),
						},
					},
				},
			},
			wantLen: 1,
			validateFn: func(hrs []orchestrator.HostResult) {
				s.Equal("db-01", hrs[0].Hostname)
				s.Equal("migrated", hrs[0].Data["stdout"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Len(tt.result.HostResults, tt.wantLen)

			if tt.validateFn != nil {
				tt.validateFn(tt.result.HostResults)
			}
		})
	}
}

func (s *ResultPublicTestSuite) TestResultsGet() {
	tests := []struct {
		name       string
		results    orchestrator.Results
		lookupName string
		wantNil    bool
		wantChange bool
	}{
		{
			name: "found",
			results: orchestrator.Results{
				"install": {Changed: true},
			},
			lookupName: "install",
			wantNil:    false,
			wantChange: true,
		},
		{
			name:       "not found",
			results:    orchestrator.Results{},
			lookupName: "missing",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.results.Get(tt.lookupName)
			if tt.wantNil {
				s.Nil(got)
			} else {
				s.Require().NotNil(got)
				s.Equal(tt.wantChange, got.Changed)
			}
		})
	}
}
