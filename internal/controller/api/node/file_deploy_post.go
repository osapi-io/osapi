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
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeFileDeploy post the node file deploy API endpoint.
func (s *Node) PostNodeFileDeploy(
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

	jobID, agentHostname, changed, err := s.JobClient.ModifyFileDeploy(
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
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileDeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileDeploy202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.FileDeployResult{
			{
				Hostname: agentHostname,
				Changed:  &changed,
			},
		},
	}, nil
}

// postNodeFileDeployBroadcast handles broadcast targets for file deploy.
func (s *Node) postNodeFileDeployBroadcast(
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
	jobID, changed, errs, err := s.JobClient.ModifyFileDeployBroadcast(
		ctx,
		target,
		objectName,
		path,
		contentType,
		mode,
		owner,
		group,
		vars,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileDeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var results []gen.FileDeployResult
	for host, c := range changed {
		c := c
		results = append(results, gen.FileDeployResult{
			Hostname: host,
			Changed:  &c,
		})
	}
	for host, errMsg := range errs {
		e := errMsg
		results = append(results, gen.FileDeployResult{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeFileDeploy202JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}
