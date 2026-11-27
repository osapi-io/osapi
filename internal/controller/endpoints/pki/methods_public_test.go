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

	"github.com/osapi-io/osapi/internal/controller/certrotation"
	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	"github.com/osapi-io/osapi/internal/controller/endpoints/pki/mocks"
	"github.com/osapi-io/osapi/internal/controller/keyrotation"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/pki"
)

type MethodsPublicTestSuite struct {
	suite.Suite
	agents *mocks.MockAgents
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.agents = mocks.NewMockAgents(gomock.NewController(s.T()))
}

func (s *MethodsPublicTestSuite) deps() pkiEndpoint.Deps {
	return pkiEndpoint.Deps{
		Agents:   s.agents,
		Hostname: "hub",
		Local: func() []pki.Cert {
			return []pki.Cert{{Name: "api", State: pki.CertOK, Component: "controller"}}
		},
	}
}

func (s *MethodsPublicTestSuite) fleet() {
	s.agents.EXPECT().ListAgents(gomock.Any()).Return([]job.AgentInfo{
		{Hostname: "web-01", Certificates: []pki.Cert{
			{Name: "agent-metrics", State: pki.CertNotMade, Component: "agent"},
		}},
		{Hostname: "web-02"},
	}, nil)
}

func (s *MethodsPublicTestSuite) TestList() {
	s.Run("the controller's certificates, then each agent's", func() {
		s.fleet()
		got, err := pkiEndpoint.List(context.Background(), s.deps(), pkiEndpoint.ListInput{})
		s.Require().NoError(err)
		s.Equal(2, got.Total)
		s.Equal("hub", got.Items[0].Hostname)
		s.Equal("api", got.Items[0].Name)
		s.Equal("web-01", got.Items[1].Hostname)
		s.Equal(pki.CertNotMade, got.Items[1].State)
	})
	s.Run("only the host that the input names", func() {
		s.fleet()
		got, err := pkiEndpoint.List(context.Background(), s.deps(),
			pkiEndpoint.ListInput{Hostname: "web-01"})
		s.Require().NoError(err)
		s.Equal(1, got.Total)
		s.Equal("agent-metrics", got.Items[0].Name)
	})
	s.Run("a controller with no certificates of its own lists the agents'", func() {
		s.fleet()
		d := s.deps()
		d.Local = nil
		got, err := pkiEndpoint.List(context.Background(), d, pkiEndpoint.ListInput{})
		s.Require().NoError(err)
		s.Equal(1, got.Total)
	})
	s.Run("a registry that cannot be read is the error", func() {
		s.agents.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("nats down"))
		_, err := pkiEndpoint.List(context.Background(), s.deps(), pkiEndpoint.ListInput{})
		s.ErrorContains(err, "nats down")
	})
}

func (s *MethodsPublicTestSuite) TestGet() {
	s.Run("with no hostname, a certificate of the controller's host", func() {
		s.fleet()
		got, err := pkiEndpoint.Get(
			context.Background(),
			s.deps(),
			pkiEndpoint.GetInput{Name: "api"},
		)
		s.Require().NoError(err)
		s.Equal("hub", got.Hostname)
	})
	s.Run("a certificate of an agent", func() {
		s.fleet()
		got, err := pkiEndpoint.Get(context.Background(), s.deps(),
			pkiEndpoint.GetInput{Name: "agent-metrics", Hostname: "web-01"})
		s.Require().NoError(err)
		s.Equal("agent", got.Component)
	})
	s.Run("a certificate that the host does not have is not found", func() {
		s.fleet()
		_, err := pkiEndpoint.Get(context.Background(), s.deps(),
			pkiEndpoint.GetInput{Name: "api", Hostname: "web-02"})
		s.ErrorIs(err, pkiEndpoint.ErrNotFound)
		s.ErrorContains(err, `host "web-02" has no certificate "api"`)
	})
	s.Run("a registry that cannot be read is the error", func() {
		s.agents.EXPECT().ListAgents(gomock.Any()).Return(nil, errors.New("nats down"))
		_, err := pkiEndpoint.Get(context.Background(), s.deps(), pkiEndpoint.GetInput{Name: "api"})
		s.ErrorContains(err, "nats down")
	})
}

func (s *MethodsPublicTestSuite) TestKeyRotation() {
	started := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	status := keyrotation.Status{
		Current:   "SHA256:old",
		Next:      "SHA256:new",
		StartedAt: started,
		SwitchBy:  started.Add(time.Hour),
		Agents:    []keyrotation.Agent{{MachineID: "m-1", Hostname: "web-01", HasNext: true}},
	}
	want := pkiEndpoint.Rotation{
		Current:   "SHA256:old",
		Next:      "SHA256:new",
		StartedAt: started,
		SwitchBy:  started.Add(time.Hour),
		Agents: []pkiEndpoint.RotationAgent{
			{Hostname: "web-01", MachineID: "m-1", HasNext: true},
		},
	}

	tests := []struct {
		name    string
		call    func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error)
		setup   func(r *mocks.MockRotator)
		want    pkiEndpoint.Rotation
		wantErr string
	}{
		{
			name: "rotate starts the rotation, and answers it",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error) {
				return pkiEndpoint.RotateKey(context.Background(), d, pkiEndpoint.RotateInput{})
			},
			setup: func(r *mocks.MockRotator) { r.EXPECT().Start(gomock.Any()).Return(status, nil) },
			want:  want,
		},
		{
			name: "a rotation that does not start",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error) {
				return pkiEndpoint.RotateKey(context.Background(), d, pkiEndpoint.RotateInput{})
			},
			setup: func(r *mocks.MockRotator) {
				r.EXPECT().Start(gomock.Any()).Return(keyrotation.Status{}, errors.New("nats down"))
			},
			wantErr: "nats down",
		},
		{
			name: "rotation shows it",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error) {
				return pkiEndpoint.KeyRotation(context.Background(), d, pkiEndpoint.RotationInput{})
			},
			setup: func(r *mocks.MockRotator) { r.EXPECT().Status(gomock.Any()).Return(status, nil) },
			want:  want,
		},
		{
			name: "with no rotation, only the current key",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error) {
				return pkiEndpoint.KeyRotation(context.Background(), d, pkiEndpoint.RotationInput{})
			},
			setup: func(r *mocks.MockRotator) {
				r.EXPECT().
					Status(gomock.Any()).
					Return(keyrotation.Status{Current: "SHA256:old"}, nil)
			},
			want: pkiEndpoint.Rotation{Current: "SHA256:old"},
		},
		{
			name: "a rotation that does not read",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.Rotation, error) {
				return pkiEndpoint.KeyRotation(context.Background(), d, pkiEndpoint.RotationInput{})
			},
			setup: func(r *mocks.MockRotator) {
				r.EXPECT().
					Status(gomock.Any()).
					Return(keyrotation.Status{}, errors.New("nats down"))
			},
			wantErr: "nats down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := mocks.NewMockRotator(gomock.NewController(s.T()))
			tc.setup(r)
			d := s.deps()
			d.Rotation = r
			got, err := tc.call(d)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *MethodsPublicTestSuite) TestNATSCertRotation() {
	started := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	status := certrotation.Status{
		Current:   "SHA256:old",
		Next:      "SHA256:new",
		StartedAt: started,
		SwitchBy:  started.Add(time.Hour),
		Agents:    []certrotation.Agent{{MachineID: "m-1", Hostname: "web-01", HasNext: true}},
	}
	want := pkiEndpoint.CertRotation{
		Current:   "SHA256:old",
		Next:      "SHA256:new",
		StartedAt: started,
		SwitchBy:  started.Add(time.Hour),
		Agents: []pkiEndpoint.CertRotationAgent{
			{Hostname: "web-01", MachineID: "m-1", HasNext: true},
		},
	}

	tests := []struct {
		name    string
		call    func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error)
		setup   func(r *mocks.MockCertRotator)
		want    pkiEndpoint.CertRotation
		wantErr string
	}{
		{
			name: "rotate starts the rotation, and answers it",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error) {
				return pkiEndpoint.RotateNATSCert(
					context.Background(),
					d,
					pkiEndpoint.CertRotateInput{},
				)
			},
			setup: func(r *mocks.MockCertRotator) { r.EXPECT().Start(gomock.Any()).Return(status, nil) },
			want:  want,
		},
		{
			name: "a rotation that does not start",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error) {
				return pkiEndpoint.RotateNATSCert(
					context.Background(),
					d,
					pkiEndpoint.CertRotateInput{},
				)
			},
			setup: func(r *mocks.MockCertRotator) {
				r.EXPECT().
					Start(gomock.Any()).
					Return(certrotation.Status{}, errors.New("nats down"))
			},
			wantErr: "nats down",
		},
		{
			name: "rotation shows it",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error) {
				return pkiEndpoint.NATSCertRotation(
					context.Background(),
					d,
					pkiEndpoint.CertRotationInput{},
				)
			},
			setup: func(r *mocks.MockCertRotator) { r.EXPECT().Status(gomock.Any()).Return(status, nil) },
			want:  want,
		},
		{
			name: "with no rotation, only the current certificate",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error) {
				return pkiEndpoint.NATSCertRotation(
					context.Background(),
					d,
					pkiEndpoint.CertRotationInput{},
				)
			},
			setup: func(r *mocks.MockCertRotator) {
				r.EXPECT().
					Status(gomock.Any()).
					Return(certrotation.Status{Current: "SHA256:old"}, nil)
			},
			want: pkiEndpoint.CertRotation{Current: "SHA256:old"},
		},
		{
			name: "a rotation that does not read",
			call: func(d pkiEndpoint.Deps) (pkiEndpoint.CertRotation, error) {
				return pkiEndpoint.NATSCertRotation(
					context.Background(),
					d,
					pkiEndpoint.CertRotationInput{},
				)
			},
			setup: func(r *mocks.MockCertRotator) {
				r.EXPECT().
					Status(gomock.Any()).
					Return(certrotation.Status{}, errors.New("nats down"))
			},
			wantErr: "nats down",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := mocks.NewMockCertRotator(gomock.NewController(s.T()))
			tc.setup(r)
			d := s.deps()
			d.CertRotation = r
			got, err := tc.call(d)
			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}
