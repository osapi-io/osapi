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

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	pkiEndpoint "github.com/osapi-io/osapi/internal/controller/endpoints/pki"
	"github.com/osapi-io/osapi/internal/controller/endpoints/pki/mocks"
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
