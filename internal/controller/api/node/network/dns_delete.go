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

package network

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// DeleteNodeNetworkDNS delete the node network DNS configuration API endpoint.
func (s *Network) DeleteNodeNetworkDNS(
	ctx context.Context,
	request gen.DeleteNodeNetworkDNSRequestObject,
) (gen.DeleteNodeNetworkDNSResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeNetworkDNS400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.DeleteNodeNetworkDNS400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	interfaceName := request.Body.InterfaceName

	s.logger.Debug("dns delete",
		slog.String("interface", interfaceName),
		slog.String("target", hostname),
	)

	data := map[string]any{"interface": interfaceName}

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeNetworkDNSBroadcast(ctx, hostname, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkDNSDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.DeleteNodeNetworkDNS200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DNSDeleteResultItem{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.DNSDeleteResultItemStatusSkipped,
					Error:    &e,
					Changed:  &falseVal,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeNetworkDNS200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSDeleteResultItem{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.DNSDeleteResultItemStatusOk,
				Changed:  &changed,
			},
		},
	}, nil
}

// deleteNodeNetworkDNSBroadcast handles broadcast targets for DNS delete.
func (s *Network) deleteNodeNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	data map[string]any,
) (gen.DeleteNodeNetworkDNSResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkDNSDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.DNSDeleteResultItem
	for host, resp := range responses {
		item := gen.DNSDeleteResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DNSDeleteResultItemStatusFailed
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		case job.StatusSkipped:
			item.Status = gen.DNSDeleteResultItemStatusSkipped
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		default:
			item.Status = gen.DNSDeleteResultItemStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeNetworkDNS200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
