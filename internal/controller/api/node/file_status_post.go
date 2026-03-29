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

package node

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeFileStatus post the node file status API endpoint.
func (s *Node) PostNodeFileStatus(
	ctx context.Context,
	request gen.PostNodeFileStatusRequestObject,
) (gen.PostNodeFileStatusResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeFileStatus400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeFileStatus400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	path := request.Body.Path
	hostname := request.Hostname

	s.logger.Debug("file status",
		slog.String("path", path),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeFileStatusBroadcast(ctx, hostname, path)
	}

	data := file.StatusRequest{Path: path}
	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"file",
		job.OperationFileStatusGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileStatus500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.PostNodeFileStatus200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.FileStatusResult{
				{
					Hostname: rawResp.Hostname,
					Error:    &e,
				},
			},
		}, nil
	}

	var result file.StatusResult
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &result)
	}

	item := gen.FileStatusResult{
		Hostname: rawResp.Hostname,
		Path:     &result.Path,
		Status:   &result.Status,
	}

	if result.SHA256 != "" {
		sha := result.SHA256
		item.Sha256 = &sha
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileStatus200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.FileStatusResult{item},
	}, nil
}

// postNodeFileStatusBroadcast handles broadcast targets for file status.
func (s *Node) postNodeFileStatusBroadcast(
	ctx context.Context,
	target string,
	path string,
) (gen.PostNodeFileStatusResponseObject, error) {
	data := file.StatusRequest{Path: path}
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"file",
		job.OperationFileStatusGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileStatus500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var items []gen.FileStatusResult
	for host, resp := range responses {
		item := gen.FileStatusResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			e := resp.Error
			item.Error = &e
		default:
			var result file.StatusResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			item.Path = &result.Path
			item.Status = &result.Status
			if result.SHA256 != "" {
				sha := result.SHA256
				item.Sha256 = &sha
			}
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileStatus200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}
