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

package certificate

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/certificate/gen"
	"github.com/retr0h/osapi/internal/job"
	certProv "github.com/retr0h/osapi/internal/provider/node/certificate"
)

// GetNodeCertificateCa lists all CA certificates on a target node.
func (s *Certificate) GetNodeCertificateCa(
	ctx context.Context,
	request gen.GetNodeCertificateCaRequestObject,
) (gen.GetNodeCertificateCaResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeCertificateCa500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("certificate ca list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeCertificateCaBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "certificate", job.OperationCertificateCAList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeCertificateCa500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeCertificateCa200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.CertificateCAEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.CertificateCAEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToCertificateCAEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeCertificateCa200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeCertificateCaBroadcast handles broadcast targets for certificate CA list.
func (s *Certificate) getNodeCertificateCaBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeCertificateCaResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"certificate",
		job.OperationCertificateCAList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeCertificateCa500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.CertificateCAEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.CertificateCAEntry{
				Hostname: h,
				Status:   gen.CertificateCAEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.CertificateCAEntry{
				Hostname: h,
				Status:   gen.CertificateCAEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToCertificateCAEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeCertificateCa200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToCertificateCAEntries converts a job response to gen CertificateCAEntry slice.
func responseToCertificateCAEntries(
	resp *job.Response,
) []gen.CertificateCAEntry {
	var entries []certProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entries)
	}

	hostname := resp.Hostname

	certs := make([]gen.CertificateCAInfo, 0, len(entries))
	for _, e := range entries {
		certs = append(certs, certificateInfoToGen(e))
	}

	return []gen.CertificateCAEntry{
		{
			Hostname:     hostname,
			Status:       gen.CertificateCAEntryStatusOk,
			Certificates: &certs,
		},
	}
}

// certificateInfoToGen converts a provider Entry to a gen CertificateCAInfo.
func certificateInfoToGen(
	e certProv.Entry,
) gen.CertificateCAInfo {
	info := gen.CertificateCAInfo{}

	if e.Name != "" {
		name := e.Name
		info.Name = &name
	}
	if e.Object != "" {
		object := e.Object
		info.Object = &object
	}
	if e.Source != "" {
		source := gen.CertificateCAInfoSource(e.Source)
		info.Source = &source
	}

	return info
}
