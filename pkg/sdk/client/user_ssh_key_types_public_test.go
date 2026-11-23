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

type UserSSHKeyTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *UserSSHKeyTypesPublicTestSuite) TestSSHKeyCollectionFromGen() {
	typ, fp, comment, opts, errMsg := "ssh-ed25519", "SHA256:abc", "alice@laptop", `from="10.0.0.0/8"`, "boom"
	jobID := openapi_types.UUID{1}

	got := client.SSHKeyCollectionFromGen([]gen.UserSshkeyKey{
		{
			Hostname:    "web-01",
			Status:      "ok",
			Type:        &typ,
			Fingerprint: &fp,
			Comment:     &comment,
			Options:     &opts,
		},
		{Hostname: "web-02", Status: "failed", Error: &errMsg},
	}, &jobID)

	suite.Equal(jobID.String(), got.JobID)
	suite.Equal([]client.SSHKeyInfoResult{
		{
			Hostname: "web-01", Status: "ok", Type: "ssh-ed25519", Fingerprint: "SHA256:abc",
			Comment: "alice@laptop", Options: `from="10.0.0.0/8"`,
		},
		{Hostname: "web-02", Status: "failed", Error: "boom"},
	}, got.Results)
	suite.Empty(client.SSHKeyCollectionFromGen(nil, nil).JobID)
}

func (suite *UserSSHKeyTypesPublicTestSuite) TestSSHKeyMutationCollectionFromGen() {
	fp := "SHA256:abc"
	changed := true

	got := client.SSHKeyMutationCollectionFromGen([]gen.UserSshkeyResult{
		{Hostname: "web-01", Status: "ok", Fingerprint: &fp, Changed: &changed},
		{Hostname: "web-02", Status: "skipped"},
	}, nil)

	suite.Equal([]client.SSHKeyMutationResult{
		{Hostname: "web-01", Status: "ok", Fingerprint: "SHA256:abc", Changed: true},
		{Hostname: "web-02", Status: "skipped"},
	}, got.Results)
}

func TestUserSSHKeyTypesPublicTestSuite(
	t *testing.T,
) {
	t.Parallel()
	suite.Run(t, new(UserSSHKeyTypesPublicTestSuite))
}
