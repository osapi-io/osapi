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

package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

type QueryPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *QueryPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	var err error
	s.jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)
}

func (s *QueryPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusSuccess() {
	hostname := "server1"
	responseData := `{
		"status": "completed",
		"data": {
			"hostname": "server1",
			"uptime": 3600000000000,
			"os_info": {"name": "Linux", "version": "5.4.0"},
			"load_averages": {"load1": 0.5, "load5": 0.3, "load15": 0.1},
			"memory_stats": {"total": 8589934592, "available": 4294967296},
			"disk_usage": [{"filesystem": "/dev/sda1", "used": 50, "available": 50}]
		}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemStatus(s.ctx, hostname)
	s.NoError(err)
	s.NotNil(result)
	s.Equal("server1", result.Hostname)
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusJobFailed() {
	hostname := "server1"
	responseData := `{
		"status": "failed",
		"error": "unable to gather system info",
		"data": {}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemStatus(s.ctx, hostname)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "job failed: unable to gather system info")
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusUnmarshalError() {
	hostname := "server1"
	responseData := `{
		"status": "completed",
		"data": "invalid_data_format"
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemStatus(s.ctx, hostname)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to unmarshal status response")
}

func (s *QueryPublicTestSuite) TestQuerySystemHostnameSuccess() {
	hostname := "server1"
	responseData := `{
		"status": "completed",
		"data": {"hostname": "server1.example.com"}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemHostname(s.ctx, hostname)
	s.NoError(err)
	s.Equal("server1.example.com", result)
}

func (s *QueryPublicTestSuite) TestQuerySystemHostnameJobFailed() {
	hostname := "server1"
	responseData := `{
		"status": "failed",
		"error": "hostname resolution failed",
		"data": {}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemHostname(s.ctx, hostname)
	s.Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "job failed: hostname resolution failed")
}

func (s *QueryPublicTestSuite) TestQuerySystemHostnameUnmarshalError() {
	hostname := "server1"
	responseData := `{
		"status": "completed",
		"data": "invalid_hostname_format"
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemHostname(s.ctx, hostname)
	s.Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "failed to unmarshal hostname response")
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSSuccess() {
	hostname := "server1"
	iface := "eth0"
	responseData := `{
		"status": "completed",
		"data": {
			"DNSServers": ["8.8.8.8", "1.1.1.1"],
			"SearchDomains": ["example.com", "local"]
		}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QueryNetworkDNS(s.ctx, hostname, iface)
	s.NoError(err)
	s.NotNil(result)
	// Verify DNS servers are populated correctly
	s.Len(result.DNSServers, 2)
	s.Equal("8.8.8.8", result.DNSServers[0])
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSJobFailed() {
	hostname := "server1"
	iface := "eth0"
	responseData := `{
		"status": "failed",
		"error": "interface not found",
		"data": {}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QueryNetworkDNS(s.ctx, hostname, iface)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "job failed: interface not found")
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSPublishError() {
	hostname := "server1"
	iface := "eth0"

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return(nil, errors.New("connection failed"))

	result, err := s.jobsClient.QueryNetworkDNS(s.ctx, hostname, iface)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to publish and wait")
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSUnmarshalError() {
	hostname := "server1"
	iface := "eth0"
	responseData := `{
		"status": "completed",
		"data": "invalid_dns_format"
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QueryNetworkDNS(s.ctx, hostname, iface)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "failed to unmarshal DNS response")
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusAny() {
	responseData := `{
		"status": "completed",
		"data": {
			"hostname": "any-server",
			"uptime": 3600000000000
		}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query._any", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemStatusAny(s.ctx)
	s.NoError(err)
	s.NotNil(result)
	s.Equal("any-server", result.Hostname)
}

func (s *QueryPublicTestSuite) TestQueryPublishErrors() {
	tests := []struct {
		name            string
		hostname        string
		expectedSubject string
		publishError    error
		queryFunc       func(string) error
	}{
		{
			name:            "QuerySystemStatus publish error",
			hostname:        "server1",
			expectedSubject: "jobs.query.server1",
			publishError:    errors.New("connection timeout"),
			queryFunc: func(hostname string) error {
				_, err := s.jobsClient.QuerySystemStatus(s.ctx, hostname)
				return err
			},
		},
		{
			name:            "QuerySystemHostname publish error",
			hostname:        "server1",
			expectedSubject: "jobs.query.server1",
			publishError:    errors.New("network unavailable"),
			queryFunc: func(hostname string) error {
				_, err := s.jobsClient.QuerySystemHostname(s.ctx, hostname)
				return err
			},
		},
		{
			name:            "QuerySystemStatusAny publish error",
			hostname:        "_any",
			expectedSubject: "jobs.query._any",
			publishError:    errors.New("no workers available"),
			queryFunc: func(_ string) error {
				_, err := s.jobsClient.QuerySystemStatusAny(s.ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockNATSClient.EXPECT().
				PublishAndWaitKV(gomock.Any(), tt.expectedSubject, gomock.Any(), s.mockKV, gomock.Any()).
				Return(nil, tt.publishError)

			err := tt.queryFunc(tt.hostname)
			s.Error(err)
			s.Contains(err.Error(), "failed to publish and wait")
		})
	}
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusWithPartialData() {
	hostname := "server1"
	responseData := `{
		"status": "completed",
		"data": {
			"hostname": "server1",
			"uptime": 3600000000000
		}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QuerySystemStatus(s.ctx, hostname)
	s.NoError(err)
	s.NotNil(result)
	s.Equal("server1", result.Hostname)
	s.Equal(time.Duration(3600000000000), result.Uptime)
	s.Nil(result.OSInfo)
	s.Nil(result.LoadAverages)
	s.Nil(result.MemoryStats)
	s.Empty(result.DiskUsage)
}

func (s *QueryPublicTestSuite) TestQueryRequestStructure() {
	tests := []struct {
		name                string
		hostname            string
		iface               string
		expectedSubject     string
		expectedType        string
		expectedCategory    string
		expectedOperation   string
		expectedDataContent map[string]interface{}
		responseData        string
		queryFunc           func() error
	}{
		{
			name:              "QuerySystemStatus request structure",
			hostname:          "server1",
			expectedSubject:   "jobs.query.server1",
			expectedType:      "query",
			expectedCategory:  "system",
			expectedOperation: "status.get",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "server1",
					"uptime": 3600000000000
				}
			}`,
			queryFunc: func() error {
				_, err := s.jobsClient.QuerySystemStatus(s.ctx, "server1")
				return err
			},
		},
		{
			name:              "QuerySystemHostname request structure",
			hostname:          "server1",
			expectedSubject:   "jobs.query.server1",
			expectedType:      "query",
			expectedCategory:  "system",
			expectedOperation: "hostname.get",
			responseData: `{
				"status": "completed",
				"data": {"hostname": "server1.example.com"}
			}`,
			queryFunc: func() error {
				_, err := s.jobsClient.QuerySystemHostname(s.ctx, "server1")
				return err
			},
		},
		{
			name:              "QueryNetworkDNS request structure",
			hostname:          "server1",
			iface:             "eth0",
			expectedSubject:   "jobs.query.server1",
			expectedType:      "query",
			expectedCategory:  "network",
			expectedOperation: "dns.get",
			expectedDataContent: map[string]interface{}{
				"interface": "eth0",
			},
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": ["8.8.8.8"],
					"SearchDomains": ["example.com"]
				}
			}`,
			queryFunc: func() error {
				_, err := s.jobsClient.QueryNetworkDNS(s.ctx, "server1", "eth0")
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockNATSClient.EXPECT().
				PublishAndWaitKV(gomock.Any(), tt.expectedSubject, gomock.Any(), s.mockKV, gomock.Any()).
				Do(func(_ context.Context, subject string, data []byte, _ interface{}, _ interface{}) {
					// Verify the subject is correct
					s.Equal(tt.expectedSubject, subject)

					// Verify the request structure
					var req map[string]interface{}
					err := json.Unmarshal(data, &req)
					s.NoError(err)
					s.Equal(tt.expectedType, req["type"])
					s.Equal(tt.expectedCategory, req["category"])
					s.Equal(tt.expectedOperation, req["operation"])

					// Verify the data payload if expected
					if tt.expectedDataContent != nil {
						var reqData map[string]interface{}

						// The data field might be stored differently depending on JSON marshaling
						switch dataField := req["data"].(type) {
						case []byte:
							err = json.Unmarshal(dataField, &reqData)
							s.NoError(err)
						case string:
							err = json.Unmarshal([]byte(dataField), &reqData)
							s.NoError(err)
						case json.RawMessage:
							err = json.Unmarshal(dataField, &reqData)
							s.NoError(err)
						case map[string]interface{}:
							// Already parsed
							reqData = dataField
						default:
							s.Failf("unexpected data field type", "got %T", dataField)
						}

						for key, expectedValue := range tt.expectedDataContent {
							s.Equal(expectedValue, reqData[key])
						}
					}
				}).
				Return([]byte(tt.responseData), nil)

			err := tt.queryFunc()
			s.NoError(err)
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryEdgeCases() {
	tests := []struct {
		name            string
		hostname        string
		iface           string
		responseData    string
		expectedSubject string
		expectError     bool
		errorContains   string
		queryFunc       func() (interface{}, error)
		validateResult  func(interface{})
	}{
		{
			name:            "QuerySystemStatus with empty hostname",
			hostname:        "",
			expectedSubject: "jobs.query.",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "default-server",
					"uptime": 1000000000000
				}
			}`,
			queryFunc: func() (interface{}, error) {
				return s.jobsClient.QuerySystemStatus(s.ctx, "")
			},
			validateResult: func(result interface{}) {
				status := result.(*job.SystemStatusResponse)
				s.Equal("default-server", status.Hostname)
				s.Equal(time.Duration(1000000000000), status.Uptime)
			},
		},
		{
			name:            "QuerySystemStatus with invalid JSON response",
			hostname:        "server1",
			expectedSubject: "jobs.query.server1",
			responseData:    `{invalid json}`,
			expectError:     true,
			errorContains:   "failed to unmarshal response",
			queryFunc: func() (interface{}, error) {
				return s.jobsClient.QuerySystemStatus(s.ctx, "server1")
			},
		},
		{
			name:            "QuerySystemHostname with empty hostname in response",
			hostname:        "server1",
			expectedSubject: "jobs.query.server1",
			responseData: `{
				"status": "completed",
				"data": {"hostname": ""}
			}`,
			queryFunc: func() (interface{}, error) {
				return s.jobsClient.QuerySystemHostname(s.ctx, "server1")
			},
			validateResult: func(result interface{}) {
				hostname := result.(string)
				s.Empty(hostname)
			},
		},
		{
			name:            "QueryNetworkDNS with empty interface",
			hostname:        "server1",
			iface:           "",
			expectedSubject: "jobs.query.server1",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": ["8.8.8.8"],
					"SearchDomains": []
				}
			}`,
			queryFunc: func() (interface{}, error) {
				return s.jobsClient.QueryNetworkDNS(s.ctx, "server1", "")
			},
			validateResult: func(result interface{}) {
				dns := result.(*dns.Config)
				s.Len(dns.DNSServers, 1)
				s.Equal("8.8.8.8", dns.DNSServers[0])
				s.Empty(dns.SearchDomains)
			},
		},
		{
			name:            "QueryNetworkDNS with empty response arrays",
			hostname:        "server1",
			iface:           "eth0",
			expectedSubject: "jobs.query.server1",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": [],
					"SearchDomains": []
				}
			}`,
			queryFunc: func() (interface{}, error) {
				return s.jobsClient.QueryNetworkDNS(s.ctx, "server1", "eth0")
			},
			validateResult: func(result interface{}) {
				dns := result.(*dns.Config)
				s.Empty(dns.DNSServers)
				s.Empty(dns.SearchDomains)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockNATSClient.EXPECT().
				PublishAndWaitKV(gomock.Any(), tt.expectedSubject, gomock.Any(), s.mockKV, gomock.Any()).
				Return([]byte(tt.responseData), nil)

			result, err := tt.queryFunc()

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorContains)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				if tt.validateResult != nil {
					tt.validateResult(result)
				}
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSWithCompleteData() {
	hostname := "server1"
	iface := "eth0"
	responseData := `{
		"status": "completed",
		"data": {
			"DNSServers": ["8.8.8.8", "1.1.1.1", "9.9.9.9"],
			"SearchDomains": ["example.com", "local", "internal"]
		}
	}`

	s.mockNATSClient.EXPECT().
		PublishAndWaitKV(gomock.Any(), "jobs.query.server1", gomock.Any(), s.mockKV, gomock.Any()).
		Return([]byte(responseData), nil)

	result, err := s.jobsClient.QueryNetworkDNS(s.ctx, hostname, iface)
	s.NoError(err)
	s.NotNil(result)
	s.Len(result.DNSServers, 3)
	s.Equal("8.8.8.8", result.DNSServers[0])
	s.Equal("1.1.1.1", result.DNSServers[1])
	s.Equal("9.9.9.9", result.DNSServers[2])
	s.Len(result.SearchDomains, 3)
	s.Equal("example.com", result.SearchDomains[0])
	s.Equal("local", result.SearchDomains[1])
	s.Equal("internal", result.SearchDomains[2])
}

func (s *QueryPublicTestSuite) TestQuerySystemStatusAll() {
	result, err := s.jobsClient.QuerySystemStatusAll(s.ctx)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "broadcast queries not yet implemented")
}

func TestQueryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(QueryPublicTestSuite))
}
