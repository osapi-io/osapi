// Copyright (c) 2026 John Dewey

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
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

type AuditTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *AuditTypesPublicTestSuite) TestAuditEntryFromGen() {
	now := time.Now().UTC().Truncate(time.Second)
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}
	operationID := "getNodeHostname"

	tests := []struct {
		name         string
		input        gen.AuditEntry
		validateFunc func(client.AuditEntry)
	}{
		{
			name: "when all fields are populated",
			input: gen.AuditEntry{
				Id:           testUUID,
				Timestamp:    now,
				User:         "admin@example.com",
				Roles:        []string{"admin", "write"},
				Method:       "GET",
				Path:         "/api/v1/node/web-01",
				ResponseCode: 200,
				DurationMs:   42,
				SourceIp:     "192.168.1.100",
				OperationId:  &operationID,
			},
			validateFunc: func(a client.AuditEntry) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", a.ID)
				suite.Equal(now, a.Timestamp)
				suite.Equal("admin@example.com", a.User)
				suite.Equal([]string{"admin", "write"}, a.Roles)
				suite.Equal("GET", a.Method)
				suite.Equal("/api/v1/node/web-01", a.Path)
				suite.Equal(200, a.ResponseCode)
				suite.Equal(int64(42), a.DurationMs)
				suite.Equal("192.168.1.100", a.SourceIP)
				suite.Equal("getNodeHostname", a.OperationID)
			},
		},
		{
			name: "when OperationId is nil",
			input: gen.AuditEntry{
				Id:           testUUID,
				Timestamp:    now,
				User:         "user@example.com",
				Roles:        []string{"read"},
				Method:       "POST",
				Path:         "/api/v1/jobs",
				ResponseCode: 201,
				DurationMs:   15,
				SourceIp:     "10.0.0.1",
				OperationId:  nil,
			},
			validateFunc: func(a client.AuditEntry) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", a.ID)
				suite.Equal(now, a.Timestamp)
				suite.Equal("user@example.com", a.User)
				suite.Equal([]string{"read"}, a.Roles)
				suite.Equal("POST", a.Method)
				suite.Equal("/api/v1/jobs", a.Path)
				suite.Equal(201, a.ResponseCode)
				suite.Equal(int64(15), a.DurationMs)
				suite.Equal("10.0.0.1", a.SourceIP)
				suite.Empty(a.OperationID)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportAuditEntryFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *AuditTypesPublicTestSuite) TestAuditListFromGen() {
	now := time.Now().UTC().Truncate(time.Second)
	testUUID1 := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x01,
	}
	testUUID2 := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x02,
	}

	tests := []struct {
		name         string
		input        *gen.ListAuditResponse
		validateFunc func(client.AuditList)
	}{
		{
			name: "when list contains items",
			input: &gen.ListAuditResponse{
				Items: []gen.AuditEntry{
					{
						Id:           testUUID1,
						Timestamp:    now,
						User:         "admin@example.com",
						Roles:        []string{"admin"},
						Method:       "GET",
						Path:         "/api/v1/health",
						ResponseCode: 200,
						DurationMs:   5,
						SourceIp:     "192.168.1.1",
					},
					{
						Id:           testUUID2,
						Timestamp:    now,
						User:         "user@example.com",
						Roles:        []string{"read"},
						Method:       "POST",
						Path:         "/api/v1/jobs",
						ResponseCode: 201,
						DurationMs:   30,
						SourceIp:     "10.0.0.1",
					},
				},
				TotalItems: 2,
			},
			validateFunc: func(al client.AuditList) {
				suite.Equal(2, al.TotalItems)
				suite.Require().Len(al.Items, 2)
				suite.Equal("550e8400-e29b-41d4-a716-446655440001", al.Items[0].ID)
				suite.Equal("admin@example.com", al.Items[0].User)
				suite.Equal("550e8400-e29b-41d4-a716-446655440002", al.Items[1].ID)
				suite.Equal("user@example.com", al.Items[1].User)
			},
		},
		{
			name: "when list is empty",
			input: &gen.ListAuditResponse{
				Items:      []gen.AuditEntry{},
				TotalItems: 0,
			},
			validateFunc: func(al client.AuditList) {
				suite.Equal(0, al.TotalItems)
				suite.Empty(al.Items)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportAuditListFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestAuditTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditTypesPublicTestSuite))
}
