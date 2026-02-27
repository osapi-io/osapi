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
	"strings"

	"github.com/retr0h/osapi/internal/api/node/gen"
)

// GetNodeDetails retrieves detailed information about a specific agent.
func (s *Node) GetNodeDetails(
	ctx context.Context,
	request gen.GetNodeDetailsRequestObject,
) (gen.GetNodeDetailsResponseObject, error) {
	agent, err := s.JobClient.GetAgent(ctx, request.Hostname)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return gen.GetNodeDetails404JSONResponse{
				Error: &errMsg,
			}, nil
		}
		return gen.GetNodeDetails500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	info := buildAgentInfo(agent)

	return gen.GetNodeDetails200JSONResponse(info), nil
}
