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

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

type UserTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *UserTypesPublicTestSuite) TestUserInfoCollectionFromGen() {
	name, home, shell, errMsg := "alice", "/home/alice", "/bin/bash", "agent unreachable"
	uid, gid := 1000, 1000
	groups := []string{"developers"}
	locked := true
	jobID := openapi_types.UUID{1}

	got := client.UserInfoCollectionFromGen([]gen.UserAccount{
		{
			Hostname: "web-01", Status: "ok", Name: &name, Uid: &uid, Gid: &gid,
			Home: &home, Shell: &shell, Groups: &groups, Locked: &locked,
		},
		{Hostname: "web-02", Status: "failed", Error: &errMsg},
	}, &jobID)

	suite.Equal(jobID.String(), got.JobID)
	suite.Equal([]client.UserInfoResult{
		{
			Hostname: "web-01", Status: "ok", Name: "alice", UID: 1000, GID: 1000,
			Home: "/home/alice", Shell: "/bin/bash", Groups: []string{"developers"}, Locked: true,
		},
		{Hostname: "web-02", Status: "failed", Error: "agent unreachable"},
	}, got.Results)
	suite.Empty(client.UserInfoCollectionFromGen(nil, nil).JobID)
}

func (suite *UserTypesPublicTestSuite) TestUserMutationCollectionFromGen() {
	name := "alice"
	changed := true

	got := client.UserMutationCollectionFromGen([]gen.UserResult{
		{Hostname: "web-01", Status: "ok", Name: &name, Changed: &changed},
		{Hostname: "web-02", Status: "skipped"},
	}, nil)

	suite.Equal([]client.UserMutationResult{
		{Hostname: "web-01", Status: "ok", Name: "alice", Changed: true},
		{Hostname: "web-02", Status: "skipped"},
	}, got.Results)
}

func TestUserTypesPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(UserTypesPublicTestSuite))
}
