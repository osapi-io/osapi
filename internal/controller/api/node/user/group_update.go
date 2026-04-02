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

// PutNodeGroup updates a group on a target node.
func (u *User) PutNodeGroup(
	ctx context.Context,
	request gen.PutNodeGroupRequestObject,
) (gen.PutNodeGroupResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeGroup400JSONResponse{Error: &errMsg}, nil
	}

	if request.Body.Members == nil {
		errMsg := "at least one field must be provided"
		return gen.PutNodeGroup400JSONResponse{Error: &errMsg}, nil
	}

	opts := userProv.UpdateGroupOpts{}
	if request.Body.Members != nil {
		opts.Members = *request.Body.Members
	}

	hostname := request.Hostname
	name := request.Name

	u.logger.Debug("group update",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := map[string]interface{}{
		"name": name,
		"opts": opts,
	}

	if job.IsBroadcastTarget(hostname) {
		return u.putNodeGroupUpdateBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := u.JobClient.Modify(
		ctx,
		hostname,
		"group",
		job.OperationGroupUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.PutNodeGroup404JSONResponse{Error: &errMsg}, nil
		}
		return gen.PutNodeGroup500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PutNodeGroup200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.GroupMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.GroupMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result userProv.GroupResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resultName := result.Name
	agentHostname := resp.Hostname

	return gen.PutNodeGroup200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.GroupMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.GroupMutationResultStatusOk,
				Name:     &resultName,
				Changed:  changed,
			},
		},
	}, nil
}

// putNodeGroupUpdateBroadcast handles broadcast targets for group update.
func (u *User) putNodeGroupUpdateBroadcast(
	ctx context.Context,
	target string,
	data map[string]interface{},
) (gen.PutNodeGroupResponseObject, error) {
	jobID, responses, err := u.JobClient.ModifyBroadcast(
		ctx,
		target,
		"group",
		job.OperationGroupUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeGroup500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.GroupMutationResult
	for host, resp := range responses {
		item := gen.GroupMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.GroupMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.GroupMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.GroupMutationResultStatusOk
			var result userProv.GroupResult
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

	return gen.PutNodeGroup200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
