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

package user

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/user/gen"
	"github.com/retr0h/osapi/internal/job"
	userProv "github.com/retr0h/osapi/internal/provider/node/user"
)

// DeleteNodeUser deletes a user on a target node.
func (u *User) DeleteNodeUser(
	ctx context.Context,
	request gen.DeleteNodeUserRequestObject,
) (gen.DeleteNodeUserResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	u.logger.Debug("user delete",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return u.deleteNodeUserBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := u.JobClient.Modify(
		ctx,
		hostname,
		"user",
		job.OperationUserDelete,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.DeleteNodeUser404JSONResponse{Error: &errMsg}, nil
		}
		return gen.DeleteNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.DeleteNodeUser200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.UserMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.UserMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result userProv.Result
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resultName := result.Name
	agentHostname := resp.Hostname

	return gen.DeleteNodeUser200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.UserMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.UserMutationResultStatusOk,
				Name:     &resultName,
				Changed:  changed,
			},
		},
	}, nil
}

// deleteNodeUserBroadcast handles broadcast targets for user delete.
func (u *User) deleteNodeUserBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.DeleteNodeUserResponseObject, error) {
	jobID, responses, err := u.JobClient.ModifyBroadcast(
		ctx,
		target,
		"user",
		job.OperationUserDelete,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.UserMutationResult
	for host, resp := range responses {
		item := gen.UserMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.UserMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.UserMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.UserMutationResultStatusOk
			var result userProv.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			resultName := result.Name
			item.Name = &resultName
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.DeleteNodeUser200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
