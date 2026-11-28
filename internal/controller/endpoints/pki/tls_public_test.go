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

package pki_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	"github.com/osapi-io/osapi/internal/controller/endpoints/pki/mocks"
	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type TLSPublicTestSuite struct {
	suite.Suite
	agents *mocks.MockAgents
}

func TestTLSPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TLSPublicTestSuite))
}

// clock is the time that the expiries of the tests are judged at.
var clock = time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)

func (s *TLSPublicTestSuite) SetupSubTest() {
	s.agents = mocks.NewMockAgents(gomock.NewController(s.T()))
	pkiEndpoint.SetNowFn(func() time.Time { return clock })
}

func (s *TLSPublicTestSuite) TearDownSubTest() {
	pkiEndpoint.ResetNowFn()
}

// off is a report of a process with FIPS 140-3 mode off and an approved
// listener.
var off = &cryptoreport.Report{
	Go:     "go1.27.3",
	Module: "v1.0.0",
	Mode:   types.FIPSOff,
	Gaps:   []string{"FIPS 140-3 mode is off: set GODEBUG=fips140=on"},
	Listeners: []cryptoreport.Listener{
		{Name: "api", TLS: true, MinVersion: "TLS 1.2", Approved: true},
	},
	Algorithms: []cryptoreport.Algorithm{cryptoreport.Seal, cryptoreport.Store},
}

// cert is an ok certificate of component that expires in days.
func cert(
	name string,
	component string,
	days int,
) pki.Cert {
	left := days

	return pki.Cert{
		Name: name, State: pki.CertOK, Component: component, Subject: "CN=hub",
		Issuer: "CN=hub", NotBefore: clock.AddDate(-1, 0, 0),
		Expires: clock.Add(time.Duration(days)*24*time.Hour + time.Hour), DaysLeft: &left,
	}
}

func (s *TLSPublicTestSuite) deps() pkiEndpoint.Deps {
	return pkiEndpoint.Deps{
		Agents:   s.agents,
		Hostname: "hub",
		Local: func() []pki.Cert {
			return []pki.Cert{cert("api", "controller", 62), cert("nats", "nats", 12)}
		},
		Processes: func(context.Context) []pkiEndpoint.Process {
			return []pkiEndpoint.Process{
				{Component: pkiEndpoint.ComponentNATS, Hostname: "hub", Crypto: off},
				{Component: pkiEndpoint.ComponentController, Hostname: "hub", Crypto: off},
			}
		},
	}
}

func (s *TLSPublicTestSuite) fleet() {
	expired := cert("agent-metrics", "agent", -3)
	expired.Expires = clock.Add(-72 * time.Hour)
	s.agents.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
		{Hostname: "web-02", Crypto: off, Certificates: []pki.Cert{
			{Name: "agent-metrics", State: pki.CertNotMade, Component: "agent"},
		}},
		{Hostname: "web-01", Crypto: off, Certificates: []pki.Cert{expired}},
	}, nil)
}

func (s *TLSPublicTestSuite) TestTLSSections() {
	tests := []struct {
		name         string
		deps         func() pkiEndpoint.Deps
		in           pkiEndpoint.TLSListInput
		setup        func()
		validateFunc func(pkiEndpoint.TLSList, error)
	}{
		{
			name: "the controller, NATS, then each agent by hostname, each with its certificates",
			deps: s.deps, setup: s.fleet,
			validateFunc: func(got pkiEndpoint.TLSList, err error) {
				s.Require().NoError(err)
				s.Require().Equal(4, got.Total)
				ctl, nats, w1, w2 := got.Items[0], got.Items[1], got.Items[2], got.Items[3]
				s.Equal(pkiEndpoint.ComponentController, ctl.Component)
				s.Equal("FIPS off · cert 62d", ctl.Summary)
				s.Equal(pkiEndpoint.ExpiryOK, ctl.Expiry)
				s.Require().Len(ctl.Certificates, 1)
				s.Equal("CN=hub", ctl.Certificates[0].Issuer)
				s.Equal(clock.AddDate(-1, 0, 0), ctl.Certificates[0].NotBefore)
				s.Equal("go1.27.3", ctl.Go)
				s.Equal("v1.0.0", ctl.Module)
				s.Len(ctl.Listeners, 1)
				s.Len(ctl.Algorithms, 2)
				s.Contains(
					ctl.Reasons,
					"FIPS 140-3 only mode refuses JetStream store: "+cryptoreport.Store.Reason,
				)

				s.Equal(pkiEndpoint.ComponentNATS, nats.Component)
				s.Equal(pkiEndpoint.ExpirySoon, nats.Expiry)
				s.Equal("FIPS off · cert 12d", nats.Summary)

				s.Equal("web-01", w1.Hostname)
				s.Equal(pkiEndpoint.ExpiryExpired, w1.Expiry)
				s.Equal("FIPS off · cert expired", w1.Summary)
				s.Equal(pkiEndpoint.ExpiryExpired, w1.Certificates[0].Expiry)

				s.Equal("web-02", w2.Hostname)
				s.Equal(pkiEndpoint.ExpiryUnknown, w2.Expiry)
				s.Equal("FIPS off", w2.Summary)
				s.Equal(pkiEndpoint.ExpiryUnknown, w2.Certificates[0].Expiry)
			},
		},
		{
			name: "only the host and the component that the input names",
			deps: s.deps, setup: s.fleet,
			in: pkiEndpoint.TLSListInput{Hostname: "hub", Component: pkiEndpoint.ComponentNATS},
			validateFunc: func(got pkiEndpoint.TLSList, err error) {
				s.Require().NoError(err)
				s.Require().Equal(1, got.Total)
				s.Equal(pkiEndpoint.ComponentNATS, got.Items[0].Component)
			},
		},
		{
			name: "a component that reports no crypto is unknown",
			deps: func() pkiEndpoint.Deps {
				d := s.deps()
				d.Local = nil
				d.Processes = nil

				return d
			},
			setup: func() {
				s.agents.EXPECT().
					ListAgents(gomock.Any()).
					Return([]job.AgentInfo{{Hostname: "web-01"}}, nil)
			},
			validateFunc: func(got pkiEndpoint.TLSList, err error) {
				s.Require().NoError(err)
				s.Require().Equal(1, got.Total)
				s.Equal("FIPS unknown", got.Items[0].Summary)
				s.Empty(got.Items[0].Reasons)
			},
		},
		{
			name: "a registry that cannot be read is the error",
			deps: s.deps,
			setup: func() {
				s.agents.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("nats down"))
			},
			validateFunc: func(_ pkiEndpoint.TLSList, err error) {
				s.ErrorContains(err, "list the agents: nats down")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			tc.validateFunc(pkiEndpoint.TLSSections(context.Background(), tc.deps(), tc.in))
		})
	}
}

func (s *TLSPublicTestSuite) TestTLSSection() {
	tests := []struct {
		name         string
		in           pkiEndpoint.TLSGetInput
		setup        func()
		validateFunc func(pkiEndpoint.TLS, error)
	}{
		{
			name:  "with no hostname, a component of the controller's host",
			in:    pkiEndpoint.TLSGetInput{Component: pkiEndpoint.ComponentController},
			setup: s.fleet,
			validateFunc: func(got pkiEndpoint.TLS, err error) {
				s.Require().NoError(err)
				s.Equal("hub", got.Hostname)
				s.Equal(pkiEndpoint.ComponentController, got.Component)
			},
		},
		{
			name: "an agent of the hostname",
			in: pkiEndpoint.TLSGetInput{
				Component: pkiEndpoint.ComponentAgent,
				Hostname:  "web-01",
			},
			setup: s.fleet,
			validateFunc: func(got pkiEndpoint.TLS, err error) {
				s.Require().NoError(err)
				s.Equal("web-01", got.Hostname)
			},
		},
		{
			name:  "a component that the host does not have is not found",
			in:    pkiEndpoint.TLSGetInput{Component: pkiEndpoint.ComponentAgent},
			setup: s.fleet,
			validateFunc: func(_ pkiEndpoint.TLS, err error) {
				s.ErrorIs(err, pkiEndpoint.ErrNoComponent)
				s.ErrorContains(err, `host "hub" has no agent`)
			},
		},
		{
			name: "a registry that cannot be read is the error",
			in:   pkiEndpoint.TLSGetInput{Component: pkiEndpoint.ComponentNATS},
			setup: func() {
				s.agents.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("nats down"))
			},
			validateFunc: func(_ pkiEndpoint.TLS, err error) {
				s.ErrorContains(err, "list the TLS sections: list the agents: nats down")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			tc.validateFunc(pkiEndpoint.TLSSection(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *TLSPublicTestSuite) TestComponentValid() {
	tests := []struct {
		name string
		c    pkiEndpoint.Component
		want bool
	}{
		{
			name: "when it is the controller, it is valid",
			c:    pkiEndpoint.ComponentController,
			want: true,
		},
		{name: "when it is NATS, it is valid", c: pkiEndpoint.ComponentNATS, want: true},
		{name: "when it is an agent, it is valid", c: pkiEndpoint.ComponentAgent, want: true},
		{name: "when it is not a component, it is not valid", c: "ui"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.c.Valid())
		})
	}
}

func (s *TLSPublicTestSuite) TestExpiryValid() {
	tests := []struct {
		name string
		e    pkiEndpoint.Expiry
		want bool
	}{
		{name: "when it is ok, it is valid", e: pkiEndpoint.ExpiryOK, want: true},
		{name: "when it is soon, it is valid", e: pkiEndpoint.ExpirySoon, want: true},
		{name: "when it is expired, it is valid", e: pkiEndpoint.ExpiryExpired, want: true},
		{name: "when it is unknown, it is valid", e: pkiEndpoint.ExpiryUnknown, want: true},
		{name: "when it is not an expiry, it is not valid", e: "later"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.e.Valid())
		})
	}
}
