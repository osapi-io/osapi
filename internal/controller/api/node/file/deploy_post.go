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

package file

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/file/gen"
	"github.com/retr0h/osapi/internal/job"
	providerFile "github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeFileDeploy post the node file deploy API endpoint.
func (s *File) PostNodeFileDeploy(
	ctx context.Context,
	request gen.PostNodeFileDeployRequestObject,
) (gen.PostNodeFileDeployResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeFileDeploy400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeFileDeploy400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	objectName := request.Body.ObjectName
	path := request.Body.Path
	contentType := string(request.Body.ContentType)

	var mode string
	if request.Body.Mode != nil {
		mode = *request.Body.Mode
	}

	var owner string
	if request.Body.Owner != nil {
		owner = *request.Body.Owner
	}

	var group string
	if request.Body.Group != nil {
		group = *request.Body.Group
	}

	var vars map[string]any
	if request.Body.Vars != nil {
		vars = *request.Body.Vars
	}

	hostname := request.Hostname

	s.logger.Debug("file deploy",
		slog.String("object_name", objectName),
		slog.String("path", path),
		slog.String("content_type", contentType),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeFileDeployBroadcast(
			ctx,
			hostname,
			objectName,
			path,
			contentType,
			mode,
			owner,
			group,
			vars,
		)
	}

	data := providerFile.DeployRequest{
		ObjectName:  objectName,
		Path:        path,
		ContentType: contentType,
		Mode:        mode,
		Owner:       owner,
		Group:       group,
		Vars:        vars,
	}
	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"file",
		job.OperationFileDeployExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileDeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		return gen.PostNodeFileDeploy202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.FileDeployResult{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.FileDeployResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileDeploy202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.FileDeployResult{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.FileDeployResultStatusOk,
				Changed:  &changed,
			},
		},
	}, nil
}

// postNodeFileDeployBroadcast handles broadcast targets for file deploy.
func (s *File) postNodeFileDeployBroadcast(
	ctx context.Context,
	target string,
	objectName string,
	path string,
	contentType string,
	mode string,
	owner string,
	group string,
	vars map[string]any,
) (gen.PostNodeFileDeployResponseObject, error) {
	data := providerFile.DeployRequest{
		ObjectName:  objectName,
		Path:        path,
		ContentType: contentType,
		Mode:        mode,
		Owner:       owner,
		Group:       group,
		Vars:        vars,
	}
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"file",
		job.OperationFileDeployExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileDeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var fileResults []gen.FileDeployResult
	for host, resp := range responses {
		item := gen.FileDeployResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.FileDeployResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.FileDeployResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.FileDeployResultStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
		}
		fileResults = append(fileResults, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileDeploy202JSONResponse{
		JobId:   &jobUUID,
		Results: fileResults,
	}, nil
}
