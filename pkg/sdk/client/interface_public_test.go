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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type InterfacePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *InterfacePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *InterfacePublicTestSuite) calls() map[string]func(c *client.Client) (any, error) {
	return map[string]func(c *client.Client) (any, error){
		"list": func(c *client.Client) (any, error) { return c.Interface.List(suite.ctx, "_any") },
		"get":  func(c *client.Client) (any, error) { return c.Interface.Get(suite.ctx, "_any", "eth0") },
		"create": func(c *client.Client) (any, error) {
			return c.Interface.Create(suite.ctx, "_any", "eth1", client.InterfaceConfigOpts{})
		},
		"update": func(c *client.Client) (any, error) {
			return c.Interface.Update(suite.ctx, "_any", "eth1", client.InterfaceConfigOpts{})
		},
		"delete": func(c *client.Client) (any, error) { return c.Interface.Delete(suite.ctx, "_any", "eth1") },
	}
}

func (suite *InterfacePublicTestSuite) TestInterface() {
	row := `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01","status":"ok",` +
		`"name":"eth0","changed":true}]}`
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		serverURL string
		wantErr   string
	}{
		{name: "when the API answers", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(row))
		}},
		{
			name:    "when the API refuses",
			wantErr: "forbidden",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
		},
		{
			name:    "when the body is empty",
			wantErr: "nil response body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name:      "when the server is unreachable",
			serverURL: "http://127.0.0.1:0",
			wantErr:   "interface",
		},
	}
	for _, tc := range tests {
		for method, fn := range suite.calls() {
			suite.Run(tc.name+": "+method, func() {
				url := tc.serverURL
				if url == "" {
					server := httptest.NewServer(tc.handler)
					defer server.Close()
					url = server.URL
				}
				resp, err := fn(client.New(url, "test-token", client.WithLogger(slog.Default())))
				if tc.wantErr != "" {
					suite.ErrorContains(err, tc.wantErr)

					return
				}
				suite.Require().NoError(err)
				suite.NotNil(resp)
			})
		}
	}
}

// TestListGroupsByHost checks that the rows of a list, one per interface,
// become one result per host, and a host that failed has no interfaces.
func (suite *InterfacePublicTestSuite) TestListGroupsByHost() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[` +
			`{"hostname":"web-01","status":"ok","name":"eth0","state":"up","mac":"52:54:00:12:34:56",` +
			`"mtu":1500,"ipv4":"10.0.0.5","ipv6":"2001:db8::5","addresses":["10.0.0.5/24"],` +
			`"dhcp4":true,"primary":true},` +
			`{"hostname":"web-01","status":"ok","name":"eth1","state":"down","mtu":1500},` +
			`{"hostname":"web-02","status":"failed","error":"boom"}]}`))
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))

	got, err := c.Interface.List(suite.ctx, "_all")

	suite.Require().NoError(err)
	suite.Equal([]client.InterfaceListResult{
		{Hostname: "web-01", Status: "ok", Interfaces: []client.InterfaceInfo{
			{
				Name: "eth0", State: "up", MAC: "52:54:00:12:34:56", MTU: 1500,
				IPv4: "10.0.0.5", IPv6: "2001:db8::5", Addresses: []string{"10.0.0.5/24"},
				DHCP4: true, Primary: true,
			},
			{Name: "eth1", State: "down", MTU: 1500},
		}},
		{Hostname: "web-02", Status: "failed", Error: "boom"},
	}, got.Data.Results)
}

func (suite *InterfacePublicTestSuite) TestGetAndMutationFields() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(
			[]byte(`{"results":[{"hostname":"web-01","status":"ok","name":"eth0","changed":true},` +
				`{"hostname":"web-02","status":"skipped","error":"unsupported"}]}`),
		)
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))

	got, err := c.Interface.Get(suite.ctx, "_all", "eth0")
	suite.Require().NoError(err)
	suite.Equal(client.InterfaceGetResult{
		Hostname: "web-01", Status: "ok", Interface: &client.InterfaceInfo{Name: "eth0"},
	}, got.Data.Results[0])
	suite.Nil(got.Data.Results[1].Interface)

	deleted, err := c.Interface.Delete(suite.ctx, "_all", "eth0")
	suite.Require().NoError(err)
	suite.Equal([]client.InterfaceMutationResult{
		{Hostname: "web-01", Status: "ok", Name: "eth0", Changed: true},
		{Hostname: "web-02", Status: "skipped", Error: "unsupported"},
	}, deleted.Data.Results)
}

// TestRequestBody checks that each option set reaches the request, and an
// option left unset does not.
func (suite *InterfacePublicTestSuite) TestRequestBody() {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = map[string]any{}
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()
	c := client.New(server.URL, "t", client.WithLogger(slog.Default()))
	yes, mtu := true, 9000

	_, err := c.Interface.Create(suite.ctx, "_any", "eth1", client.InterfaceConfigOpts{
		DHCP4: &yes, DHCP6: &yes, Addresses: []string{"10.0.1.5/24"}, Gateway4: "10.0.1.1",
		Gateway6: "2001:db8::1", MTU: &mtu, MACAddress: "52:54:00:aa:bb:cc", WakeOnLAN: &yes,
	})
	suite.Require().NoError(err)
	suite.Equal(map[string]any{
		"dhcp4": true, "dhcp6": true, "addresses": []any{"10.0.1.5/24"}, "gateway4": "10.0.1.1",
		"gateway6": "2001:db8::1", "mtu": float64(9000), "mac_address": "52:54:00:aa:bb:cc",
		"wakeonlan": true,
	}, body)

	_, err = c.Interface.Update(suite.ctx, "_any", "eth1", client.InterfaceConfigOpts{})
	suite.Require().NoError(err)
	suite.Empty(body)
}

func TestInterfacePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(InterfacePublicTestSuite))
}
