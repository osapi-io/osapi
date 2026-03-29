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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeFileUndeploy post the node file undeploy API endpoint.
func (s *Node) PostNodeFileUndeploy(
	ctx context.Context,
	request gen.PostNodeFileUndeployRequestObject,
) (gen.PostNodeFileUndeployResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeFileUndeploy400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeFileUndeploy400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	path := request.Body.Path
	hostname := request.Hostname

	s.logger.Debug("file undeploy",
		slog.String("path", path),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeFileUndeployBroadcast(ctx, hostname, path)
	}

	data := file.UndeployRequest{Path: path}
	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"file",
		job.OperationFileUndeployExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileUndeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		return gen.PostNodeFileUndeploy202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.FileUndeployResult{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.FileUndeployResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileUndeploy202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.FileUndeployResult{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.FileUndeployResultStatusOk,
				Changed:  &changed,
			},
		},
	}, nil
}

// postNodeFileUndeployBroadcast handles broadcast targets for file undeploy.
func (s *Node) postNodeFileUndeployBroadcast(
	ctx context.Context,
	target string,
	path string,
) (gen.PostNodeFileUndeployResponseObject, error) {
	data := file.UndeployRequest{Path: path}
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"file",
		job.OperationFileUndeployExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileUndeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var fileResults []gen.FileUndeployResult
	for host, resp := range responses {
		item := gen.FileUndeployResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.FileUndeployResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.FileUndeployResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.FileUndeployResultStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
		}
		fileResults = append(fileResults, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileUndeploy202JSONResponse{
		JobId:   &jobUUID,
		Results: fileResults,
	}, nil
}
