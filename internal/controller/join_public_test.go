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

package controller_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/pki"
)

type JoinUserPublicTestSuite struct {
	suite.Suite
}

func TestJoinUserPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(JoinUserPublicTestSuite))
}

func (s *JoinUserPublicTestSuite) TestJoinAgentUser() {
	s.Run("static users give each agent the agent user", func() {
		var cfg config.Config
		cfg.Controller.Join.Agent = config.NATSUser{Username: "agent", Password: "ap"}
		user, err := controller.ExportJoinAgentUser(cfg, "osapi")
		s.Require().NoError(err)
		got, err := user(pki.JoinRequest{})
		s.Require().NoError(err)
		s.Equal(pki.JoinNATS{Username: "agent", Password: "ap"}, got)
	})

	s.Run("operator mode mints a JWT for the agent's own key", func() {
		acct, _ := nkeys.CreateAccount()
		acctPub, _ := acct.PublicKey()
		seed, _ := acct.Seed()
		file := filepath.Join(s.T().TempDir(), "signing.nk")
		s.Require().NoError(os.WriteFile(file, seed, 0o600))
		var cfg config.Config
		cfg.Controller.Join.SigningKeyFile = file
		cfg.Controller.Join.Account = acctPub
		cfg.NATS.Stream = config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}
		user, err := controller.ExportJoinAgentUser(cfg, "osapi")
		s.Require().NoError(err)

		kp, _ := nkeys.CreateUser()
		pub, _ := kp.PublicKey()
		got, err := user(pki.JoinRequest{MachineID: "m-1", NATSUserKey: pub})
		s.Require().NoError(err)
		uc, err := jwt.DecodeUserClaims(got.JWT)
		s.Require().NoError(err)
		s.Equal(pub, uc.Subject)
		s.Equal("m-1", uc.Name)
		s.Contains(uc.Pub.Allow, "osapi.enroll.request", "the agent's permissions")
		s.Empty(got.Password)

		_, err = user(pki.JoinRequest{MachineID: "m-1"})
		s.ErrorContains(err, "sent no NATS user key")

		_, err = user(pki.JoinRequest{MachineID: "m-1", NATSUserKey: "bad"})
		s.ErrorContains(err, "is not a user public key")
	})
}
