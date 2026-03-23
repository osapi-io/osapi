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

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
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

	jobID, agentHostname, changed, err := s.JobClient.ModifyFileUndeploy(
		ctx,
		hostname,
		path,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileUndeploy500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PostNodeFileUndeploy202JSONResponse{
		JobId:    jobID,
		Hostname: agentHostname,
		Changed:  changed,
	}, nil
}
