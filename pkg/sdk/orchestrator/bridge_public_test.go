package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	client "github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type BridgePublicTestSuite struct {
	suite.Suite
}

func TestBridgePublicTestSuite(t *testing.T) {
	suite.Run(t, new(BridgePublicTestSuite))
}

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type testNested struct {
	Label string     `json:"label"`
	Inner testStruct `json:"inner"`
}

func (s *BridgePublicTestSuite) TestStructToMap() {
	tests := []struct {
		name       string
		input      any
		validateFn func(m map[string]any)
	}{
		{
			name:  "converts struct with json tags to map",
			input: testStruct{Name: "web-01", Value: 42},
			validateFn: func(m map[string]any) {
				s.Require().NotNil(m)
				s.Equal("web-01", m["name"])
				s.Equal(float64(42), m["value"])
			},
		},
		{
			name:  "returns nil for nil input",
			input: nil,
			validateFn: func(m map[string]any) {
				s.Nil(m)
			},
		},
		{
			name: "handles nested structs",
			input: testNested{
				Label: "parent",
				Inner: testStruct{Name: "child", Value: 7},
			},
			validateFn: func(m map[string]any) {
				s.Require().NotNil(m)
				s.Equal("parent", m["label"])

				inner, ok := m["inner"].(map[string]any)
				s.Require().True(ok)
				s.Equal("child", inner["name"])
				s.Equal(float64(7), inner["value"])
			},
		},
		{
			name: "converts struct without json tags using field names",
			input: client.HostnameResult{
				Hostname: "web-01",
				Changed:  true,
			},
			validateFn: func(m map[string]any) {
				s.Require().NotNil(m)
				s.Equal("web-01", m["Hostname"])
				s.Equal(true, m["Changed"])
			},
		},
		{
			name:  "returns nil for unmarshalable input",
			input: make(chan int),
			validateFn: func(m map[string]any) {
				s.Nil(m)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := orchestrator.StructToMap(tt.input)
			tt.validateFn(got)
		})
	}
}

func (s *BridgePublicTestSuite) TestCollectionResult() {
	mapper := func(r client.HostnameResult) orchestrator.HostResult {
		return orchestrator.HostResult{
			Hostname: r.Hostname,
			Changed:  r.Changed,
			Error:    r.Error,
		}
	}

	tests := []struct {
		name       string
		col        client.Collection[client.HostnameResult]
		toHost     func(client.HostnameResult) orchestrator.HostResult
		validateFn func(result *orchestrator.Result)
	}{
		{
			name: "single result with auto-populated data",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{
					{Hostname: "web-01", Changed: false},
				},
				JobID: "job-123",
			},
			toHost: mapper,
			validateFn: func(result *orchestrator.Result) {
				s.Equal("job-123", result.JobID)
				s.False(result.Changed)
				s.Require().Len(result.HostResults, 1)

				hr := result.HostResults[0]
				s.Equal("web-01", hr.Hostname)
				s.False(hr.Changed)
				s.Require().NotNil(hr.Data, "Data should be auto-populated via StructToMap")
				s.Equal("web-01", hr.Data["Hostname"])
			},
		},
		{
			name: "multiple results with changed true when any host changed",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{
					{Hostname: "web-01", Changed: false},
					{Hostname: "web-02", Changed: true},
				},
				JobID: "job-456",
			},
			toHost: mapper,
			validateFn: func(result *orchestrator.Result) {
				s.Equal("job-456", result.JobID)
				s.True(result.Changed)
				s.Len(result.HostResults, 2)
				s.False(result.HostResults[0].Changed)
				s.True(result.HostResults[1].Changed)
			},
		},
		{
			name: "empty results returns result with empty host results",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{},
				JobID:   "job-789",
			},
			toHost: mapper,
			validateFn: func(result *orchestrator.Result) {
				s.Equal("job-789", result.JobID)
				s.False(result.Changed)
				s.Empty(result.HostResults)
			},
		},
		{
			name: "data auto-populated via StructToMap when mapper leaves it nil",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{
					{Hostname: "db-01", Changed: false, Error: "timeout"},
				},
				JobID: "job-auto",
			},
			toHost: func(r client.HostnameResult) orchestrator.HostResult {
				return orchestrator.HostResult{
					Hostname: r.Hostname,
					Changed:  r.Changed,
					Error:    r.Error,
					// Data intentionally left nil
				}
			},
			validateFn: func(result *orchestrator.Result) {
				hr := result.HostResults[0]
				s.Require().NotNil(hr.Data)
				s.Equal("db-01", hr.Data["Hostname"])
				s.Equal("timeout", hr.Data["Error"])
			},
		},
		{
			name: "data preserved when mapper sets it explicitly",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{
					{Hostname: "app-01", Changed: true},
				},
				JobID: "job-explicit",
			},
			toHost: func(r client.HostnameResult) orchestrator.HostResult {
				return orchestrator.HostResult{
					Hostname: r.Hostname,
					Changed:  r.Changed,
					Data:     map[string]any{"custom": "value"},
				}
			},
			validateFn: func(result *orchestrator.Result) {
				hr := result.HostResults[0]
				s.Require().NotNil(hr.Data)
				s.Equal("value", hr.Data["custom"])
				// Should NOT contain auto-populated fields
				_, hasHostname := hr.Data["Hostname"]
				s.False(hasHostname, "mapper-set Data should not be overwritten")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := orchestrator.CollectionResult(tt.col, tt.toHost)
			s.Require().NotNil(result)
			tt.validateFn(result)
		})
	}
}
