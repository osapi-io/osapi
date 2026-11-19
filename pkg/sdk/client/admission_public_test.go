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
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type AdmissionPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func TestAdmissionPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AdmissionPublicTestSuite))
}

func (suite *AdmissionPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

// sut is a client of a server that answers with status and body, and
// records the request it got. A status of 0 is a server that is not there.
func (suite *AdmissionPublicTestSuite) sut(
	status int,
	body string,
	got *http.Request,
	gotBody *string,
) *client.Client {
	if status == 0 {
		return client.New("http://127.0.0.1:0", "test-token", client.WithLogger(slog.Default()))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got != nil {
			*got = *r.Clone(context.Background())
		}
		if gotBody != nil {
			b, _ := io.ReadAll(r.Body)
			*gotBody = string(b)
		}
		if body != "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	suite.T().Cleanup(server.Close)

	return client.New(server.URL, "test-token", client.WithLogger(slog.Default()))
}

const record = `{"machine_id":"m-1","hostname":"web-01","fingerprint":"SHA256:ab",` +
	`"state":"pending","online":true,"last_seen":"2026-11-19T01:00:00Z",` +
	`"previous":"revoked","previous_fingerprint":"SHA256:old","reason":"r"}`

func (suite *AdmissionPublicTestSuite) TestList() {
	suite.Run("when the server has records returns each, with the state filter sent", func() {
		var req http.Request
		resp, err := suite.sut(http.StatusOK, `{"items":[`+record+`],"total_items":1}`, &req, nil).
			Admission.List(suite.ctx, "pending")

		suite.Require().NoError(err)
		suite.Equal("pending", req.URL.Query().Get("state"))
		suite.Equal(1, resp.Data.TotalItems)
		a := resp.Data.Items[0]
		suite.Equal("m-1", a.MachineID)
		suite.True(a.Online)
		suite.NotNil(a.LastSeen)
		suite.Equal("revoked", a.Previous)
		suite.Equal("SHA256:old", a.PreviousFingerprint)
		suite.Equal("r", a.Reason)
	})

	suite.Run("when no state is given sends none", func() {
		var req http.Request
		_, err := suite.sut(http.StatusOK, `{"items":[],"total_items":0}`, &req, nil).
			Admission.List(suite.ctx, "")

		suite.Require().NoError(err)
		suite.False(req.URL.Query().Has("state"))
	})

	suite.Run("when PKI is off returns its 503 message", func() {
		_, err := suite.sut(
			http.StatusServiceUnavailable,
			`{"error":"set controller.pki.enabled"}`,
			nil,
			nil,
		).
			Admission.List(
			suite.ctx,
			"",
		)

		var target *client.UnexpectedStatusError
		suite.Require().ErrorAs(err, &target)
		suite.Contains(err.Error(), "controller.pki.enabled")
	})

	suite.Run("when the body is empty returns UnexpectedStatusError", func() {
		_, err := suite.sut(http.StatusOK, "", nil, nil).Admission.List(suite.ctx, "")

		var target *client.UnexpectedStatusError
		suite.Require().ErrorAs(err, &target)
		suite.Contains(target.Message, "nil response body")
	})

	suite.Run("when the request fails wraps the error", func() {
		_, err := suite.sut(0, "", nil, nil).Admission.List(suite.ctx, "")

		suite.ErrorContains(err, "list admission records:")
	})
}

func (suite *AdmissionPublicTestSuite) TestGet() {
	suite.Run("when the machine has a record returns it", func() {
		var req http.Request
		resp, err := suite.sut(http.StatusOK, record, &req, nil).Admission.Get(suite.ctx, "m-1")

		suite.Require().NoError(err)
		suite.Equal("/api/admission/m-1", req.URL.Path)
		suite.Equal("web-01", resp.Data.Hostname)
	})

	suite.Run("when the machine has no record returns NotFoundError", func() {
		_, err := suite.sut(http.StatusNotFound, `{"error":"no record"}`, nil, nil).
			Admission.Get(suite.ctx, "m-x")

		var target *client.NotFoundError
		suite.ErrorAs(err, &target)
	})

	suite.Run("when the body is empty returns UnexpectedStatusError", func() {
		_, err := suite.sut(http.StatusOK, "", nil, nil).Admission.Get(suite.ctx, "m-1")

		var target *client.UnexpectedStatusError
		suite.Require().ErrorAs(err, &target)
		suite.Contains(target.Message, "nil response body")
	})

	suite.Run("when the request fails wraps the error", func() {
		_, err := suite.sut(0, "", nil, nil).Admission.Get(suite.ctx, "m-1")

		suite.ErrorContains(err, "get admission record m-1:")
	})
}

func (suite *AdmissionPublicTestSuite) TestDecisions() {
	type call func(c *client.Client) error
	admit := func(c *client.Client) error { return c.Admission.Admit(suite.ctx, "m-1", "SHA256:ab") }
	reject := func(c *client.Client) error { return c.Admission.Reject(suite.ctx, "m-1", "unknown") }
	revoke := func(c *client.Client) error { return c.Admission.Revoke(suite.ctx, "m-1") }

	tests := []struct {
		name string
		call call
		path string
		body map[string]string
		fail string
	}{
		{
			name: "admit", call: admit, path: "/api/admission/m-1/admit",
			body: map[string]string{"fingerprint": "SHA256:ab"}, fail: "admit machine m-1:",
		},
		{
			name: "reject", call: reject, path: "/api/admission/m-1/reject",
			body: map[string]string{"reason": "unknown"}, fail: "reject machine m-1:",
		},
		{
			name: "revoke",
			call: revoke,
			path: "/api/admission/m-1/revoke",
			fail: "revoke machine m-1:",
		},
	}

	for _, tc := range tests {
		suite.Run("when "+tc.name+" succeeds sends its request", func() {
			var req http.Request
			var body string
			suite.Require().NoError(tc.call(suite.sut(http.StatusNoContent, "", &req, &body)))

			suite.Equal(http.MethodPost, req.Method)
			suite.Equal(tc.path, req.URL.Path)
			if tc.body != nil {
				var got map[string]string
				suite.Require().NoError(json.Unmarshal([]byte(body), &got))
				suite.Equal(tc.body, got)
			}
		})

		suite.Run("when "+tc.name+" finds no machine returns NotFoundError", func() {
			err := tc.call(suite.sut(http.StatusNotFound, `{"error":"no machine"}`, nil, nil))

			var target *client.NotFoundError
			suite.ErrorAs(err, &target)
		})

		suite.Run("when "+tc.name+"'s request fails wraps the error", func() {
			suite.ErrorContains(tc.call(suite.sut(0, "", nil, nil)), tc.fail)
		})
	}

	suite.Run("when admit's fingerprint differs returns ConflictError", func() {
		err := admit(suite.sut(http.StatusConflict, `{"error":"fingerprint differs"}`, nil, nil))

		var target *client.ConflictError
		suite.ErrorAs(err, &target)
	})

	suite.Run("when a decision is refused returns the error", func() {
		err := revoke(suite.sut(http.StatusForbidden, `{"error":"forbidden"}`, nil, nil))

		suite.True(errors.As(err, new(*client.AuthError)))
	})
}
