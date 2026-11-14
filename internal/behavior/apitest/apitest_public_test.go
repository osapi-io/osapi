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

package apitest_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/apitest"
	"github.com/osapi-io/osapi/internal/behaviors/sysctl"
	"github.com/osapi-io/osapi/internal/behaviors/sysctl/gen"
)

// ApitestPublicTestSuite runs one case of each kind through sysctl's
// generated routes, the routes every generated test uses the harness on.
type ApitestPublicTestSuite struct {
	suite.Suite

	api *apitest.Harness
}

func (s *ApitestPublicTestSuite) SetupSuite() {
	s.api = apitest.New(s.T(), gen.Routes)
}

func (s *ApitestPublicTestSuite) TestRun() {
	get := apitest.Request{
		Method: http.MethodGet,
		Path:   "/api/next/node/server1/sysctl/vm.swappiness",
	}
	key := sysctl.GetInput{Key: "vm.swappiness"}

	s.api.Run(s.T(), []apitest.Case{
		{
			Name:    "when a read completes, its body is the host's row",
			Request: get,
			Token:   apitest.Read,
			Job: apitest.Query(
				"server1",
				"sysctl.get",
				key,
				apitest.Completed("server1", `{"key":"vm.swappiness","value":"10"}`),
			),
			WantCode: http.StatusOK,
			WantBody: `{"job_id":"job-1","results":[{"hostname":"server1","key":"vm.swappiness","status":"ok","value":"10"}]}`,
		},
		{
			Name: "when a write is broadcast, every reply is a row in hostname order",
			Request: apitest.Request{
				Method: http.MethodPut,
				Path:   "/api/next/node/_all/sysctl/vm.swappiness",
				Body:   `{"value":"10"}`,
			},
			Token: apitest.Write,
			Job: apitest.Modify("_all", "sysctl.update",
				sysctl.UpdateInput{Key: "vm.swappiness", Value: "10"},
				apitest.TimedOut("server2"),
				apitest.Completed("server1", `{"key":"vm.swappiness","changed":true}`),
				apitest.Failed("server3", "boom"),
			),
			WantCode: http.StatusOK,
			WantBody: `{"job_id":"job-1","results":[` +
				`{"changed":true,"hostname":"server1","key":"vm.swappiness","status":"ok"},` +
				`{"error":"","hostname":"server2","status":"timeout"},` +
				`{"error":"boom","hostname":"server3","status":"failed"}]}`,
		},
		{
			Name: "when a write completes on one host, its row says whether it changed",
			Request: apitest.Request{
				Method: http.MethodDelete,
				Path:   "/api/next/node/server1/sysctl/vm.swappiness",
			},
			Token: apitest.Write,
			Job: apitest.Modify(
				"server1",
				"sysctl.delete",
				sysctl.DeleteInput{Key: "vm.swappiness"},
				apitest.Completed("server1", `{"key":"vm.swappiness","changed":false}`),
			),
			WantCode: http.StatusOK,
			WantBody: `{"job_id":"job-1","results":[{"changed":false,"hostname":"server1","key":"vm.swappiness","status":"ok"}]}`,
		},
		{
			Name:    "when a broadcast read fails, it is a 500",
			Request: apitest.Request{Method: http.MethodGet, Path: "/api/next/node/_all/sysctl"},
			Token:   apitest.Read,
			Job: apitest.Query("_all", "sysctl.list", sysctl.ListInput{}).
				Fails(apitest.ErrUnreachable),
			WantCode: http.StatusInternalServerError,
		},
		{
			Name:    "when the host does not have it, it is a 404",
			Request: get,
			Token:   apitest.Read,
			Job: apitest.Query("server1", "sysctl.get", key).
				Fails(apitest.NotFound("server1")),
			WantCode: http.StatusNotFound,
		},
		{
			Name:     "when the host skips it, its row says skipped",
			Request:  get,
			Token:    apitest.Read,
			Job:      apitest.Query("server1", "sysctl.get", key, apitest.Skipped("server1")),
			WantCode: http.StatusOK,
			WantBody: `{"job_id":"job-1","results":[{"error":"unsupported","hostname":"server1","status":"skipped"}]}`,
		},
		{
			Name:     "when there is no token, no job is queued",
			Request:  get,
			Token:    apitest.NoToken,
			WantCode: http.StatusUnauthorized,
		},
		{
			Name:     "when the token has the admin role, the request runs",
			Request:  get,
			Token:    apitest.Admin,
			Job:      apitest.Query("server1", "sysctl.get", key, apitest.Skipped("server1")),
			WantCode: http.StatusOK,
		},
		{
			Name:     "when the token grants nothing, it is a 403",
			Request:  get,
			Token:    apitest.Lacking,
			WantCode: http.StatusForbidden,
		},
	})
}

func TestApitestPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ApitestPublicTestSuite))
}
