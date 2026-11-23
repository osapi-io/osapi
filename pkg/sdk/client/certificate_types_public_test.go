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

type CertificateTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *CertificateTypesPublicTestSuite) TestCertificateInfoCollectionFromGen() {
	name, errMsg, path, source := "internal-ca", "agent unreachable", "/usr/local/share/ca-certificates/osapi-internal-ca.crt", "local"
	subject, fp, notAfter, object := "CN=internal", "6513", "2126-10-29T23:38:01Z", "internal-ca.pem"
	yes := true
	jobID := openapi_types.UUID{1}

	got := client.CertificateInfoCollectionFromGen([]gen.CertificateEntry{
		{
			Hostname:    "web-01",
			Status:      "ok",
			Name:        &name,
			Path:        &path,
			Source:      &source,
			Enabled:     &yes,
			Subject:     &subject,
			Fingerprint: &fp,
			NotAfter:    &notAfter,
			Managed:     &yes,
			Object:      &object,
		},
		{Hostname: "web-02", Status: "failed", Error: &errMsg},
	}, &jobID)

	suite.Equal(jobID.String(), got.JobID)
	suite.Equal([]client.CertificateInfoResult{
		{
			Hostname:    "web-01",
			Status:      "ok",
			Name:        "internal-ca",
			Path:        path,
			Source:      "local",
			Enabled:     true,
			Subject:     subject,
			Fingerprint: fp,
			NotAfter:    notAfter,
			Managed:     true,
			Object:      object,
		},
		{Hostname: "web-02", Status: "failed", Error: "agent unreachable"},
	}, got.Results)
	suite.Empty(client.CertificateInfoCollectionFromGen(nil, nil).JobID)
}

func (suite *CertificateTypesPublicTestSuite) TestCertificateMutationCollectionFromGen() {
	name := "internal-ca"
	changed := true

	got := client.CertificateMutationCollectionFromGen([]gen.CertificateResult{
		{Hostname: "web-01", Status: "ok", Name: &name, Changed: &changed},
		{Hostname: "web-02", Status: "skipped"},
	}, nil)

	suite.Equal([]client.CertificateMutationResult{
		{Hostname: "web-01", Status: "ok", Name: "internal-ca", Changed: true},
		{Hostname: "web-02", Status: "skipped"},
	}, got.Results)
}

func TestCertificateTypesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CertificateTypesPublicTestSuite))
}
