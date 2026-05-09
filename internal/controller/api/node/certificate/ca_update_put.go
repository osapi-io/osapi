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
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/certificate/gen"
	"github.com/retr0h/osapi/internal/job"
	certProv "github.com/retr0h/osapi/internal/provider/node/certificate"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNodeCertificateCa updates a CA certificate on a target node.
func (s *Certificate) PutNodeCertificateCa(
	ctx context.Context,
	request gen.PutNodeCertificateCaRequestObject,
) (gen.PutNodeCertificateCaResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeCertificateCa400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeCertificateCa400JSONResponse{Error: &errMsg}, nil
	}

	entry := certProv.Entry{
		Name:   request.Name,
		Object: request.Body.Object,
	}

	hostname := request.Hostname

	s.logger.Debug(
		"certificate ca update",
		slog.String("target", hostname),
		slog.String("name", entry.Name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeCertificateCaUpdateBroadcast(ctx, hostname, entry)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"certificate",
		job.OperationCertificateCAUpdate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.PutNodeCertificateCa404JSONResponse{Error: &errMsg}, nil
		}
		return gen.PutNodeCertificateCa500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PutNodeCertificateCa200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.CertificateCAMutationEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.CertificateCAMutationEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result certProv.UpdateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	name := result.Name
	agentHostname := resp.Hostname

	return gen.PutNodeCertificateCa200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.CertificateCAMutationEntry{
			{
				Hostname: agentHostname,
				Status:   gen.CertificateCAMutationEntryStatusOk,
				Name:     &name,
				Changed:  changed,
			},
		},
	}, nil
}

// putNodeCertificateCaUpdateBroadcast handles broadcast targets for certificate CA update.
func (s *Certificate) putNodeCertificateCaUpdateBroadcast(
	ctx context.Context,
	target string,
	entry certProv.Entry,
) (gen.PutNodeCertificateCaResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"certificate",
		job.OperationCertificateCAUpdate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeCertificateCa500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.CertificateCAMutationEntry
	for host, resp := range responses {
		item := gen.CertificateCAMutationEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.CertificateCAMutationEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.CertificateCAMutationEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.CertificateCAMutationEntryStatusOk
			var result certProv.UpdateResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			name := result.Name
			item.Name = &name
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PutNodeCertificateCa200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
