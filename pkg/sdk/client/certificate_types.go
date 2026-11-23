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

package client

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// CertificateInfoResult is one certificate authority on one host, from a
// list or a get. A host that failed or skipped has a row with its status and
// no CA.
type CertificateInfoResult struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Source      string `json:"source,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	NotAfter    string `json:"not_after,omitempty"`
	Managed     bool   `json:"managed,omitempty"`
	Object      string `json:"object,omitempty"`
	Error       string `json:"error,omitempty"`
}

// CertificateMutationResult is what a create, an update or a delete did on
// one host.
type CertificateMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// CertificateCreateOpts is the CA to add.
type CertificateCreateOpts struct {
	// Name is the name of the CA, such as internal-ca. Required.
	Name string
	// Object is the object of the Object Store that has the PEM. Required.
	Object string
}

// CertificateUpdateOpts is what to write a CA that osapi deployed from.
type CertificateUpdateOpts struct {
	// Object is the object that has the PEM. When empty, the CA is written
	// from the object that it was deployed from.
	Object string
}

// certificateInfoCollectionFromGen converts the rows of a list or a get.
func certificateInfoCollectionFromGen(
	rows []gen.CertificateEntry,
	jobID *openapi_types.UUID,
) Collection[CertificateInfoResult] {
	results := make([]CertificateInfoResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, CertificateInfoResult{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Name:        derefString(r.Name),
			Path:        derefString(r.Path),
			Source:      derefString(r.Source),
			Enabled:     derefBool(r.Enabled),
			Subject:     derefString(r.Subject),
			Fingerprint: derefString(r.Fingerprint),
			NotAfter:    derefString(r.NotAfter),
			Managed:     derefBool(r.Managed),
			Object:      derefString(r.Object),
			Error:       derefString(r.Error),
		})
	}

	return Collection[CertificateInfoResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}

// certificateMutationCollectionFromGen converts the rows of a create, an update or
// a delete.
func certificateMutationCollectionFromGen(
	rows []gen.CertificateResult,
	jobID *openapi_types.UUID,
) Collection[CertificateMutationResult] {
	results := make([]CertificateMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, CertificateMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[CertificateMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(jobID),
	}
}
