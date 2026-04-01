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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// CertificateCAResult represents a CA certificate list result from a single agent.
type CertificateCAResult struct {
	Hostname     string          `json:"hostname"`
	Status       string          `json:"status"`
	Certificates []CertificateCA `json:"certificates,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// CertificateCA represents a single CA certificate entry.
type CertificateCA struct {
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
	Object string `json:"object,omitempty"`
}

// CertificateCAMutationResult represents the result of a CA certificate
// create, update, or delete operation.
type CertificateCAMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// CertificateCreateOpts contains options for creating a CA certificate.
type CertificateCreateOpts struct {
	// Name is the certificate name (required).
	Name string
	// Object is the object store reference for the PEM file (required).
	Object string
}

// CertificateUpdateOpts contains options for updating a CA certificate.
type CertificateUpdateOpts struct {
	// Object is the new object store reference for the PEM file (required).
	Object string
}

// certificateCACollectionFromGen converts a gen.CertificateCACollectionResponse
// to a Collection[CertificateCAResult].
func certificateCACollectionFromGen(
	g *gen.CertificateCACollectionResponse,
) Collection[CertificateCAResult] {
	results := make([]CertificateCAResult, 0, len(g.Results))
	for _, r := range g.Results {
		result := CertificateCAResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Certificates != nil {
			certs := make([]CertificateCA, 0, len(*r.Certificates))
			for _, c := range *r.Certificates {
				certs = append(certs, certificateCAInfoFromGen(c))
			}
			result.Certificates = certs
		}

		results = append(results, result)
	}

	return Collection[CertificateCAResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// certificateCAInfoFromGen converts a gen.CertificateCAInfo to a CertificateCA.
func certificateCAInfoFromGen(
	c gen.CertificateCAInfo,
) CertificateCA {
	var source string
	if c.Source != nil {
		source = string(*c.Source)
	}

	return CertificateCA{
		Name:   derefString(c.Name),
		Source: source,
		Object: derefString(c.Object),
	}
}

// certificateCAMutationCollectionFromGen converts a gen.CertificateCAMutationResponse
// to a Collection[CertificateCAMutationResult].
func certificateCAMutationCollectionFromGen(
	g *gen.CertificateCAMutationResponse,
) Collection[CertificateCAMutationResult] {
	results := make([]CertificateCAMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, certificateCAMutationResultFromGen(r))
	}

	return Collection[CertificateCAMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// certificateCAMutationResultFromGen converts a gen.CertificateCAMutationEntry
// to a CertificateCAMutationResult.
func certificateCAMutationResultFromGen(
	r gen.CertificateCAMutationEntry,
) CertificateCAMutationResult {
	return CertificateCAMutationResult{
		Hostname: r.Hostname,
		Status:   string(r.Status),
		Name:     derefString(r.Name),
		Changed:  derefBool(r.Changed),
		Error:    derefString(r.Error),
	}
}
