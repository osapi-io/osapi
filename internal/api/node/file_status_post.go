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

	"github.com/retr0h/osapi/internal/api/node/gen"
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

	jobID, result, agentHostname, err := s.JobClient.QueryFileStatus(
		ctx,
		hostname,
		path,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeFileStatus500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	changed := false
	resp := gen.PostNodeFileStatus200JSONResponse{
		JobId:    jobID,
		Hostname: agentHostname,
		Path:     result.Path,
		Status:   result.Status,
		Changed:  &changed,
	}

	if result.SHA256 != "" {
		sha := result.SHA256
		resp.Sha256 = &sha
	}

	return resp, nil
}
