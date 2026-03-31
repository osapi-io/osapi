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

// PutNodeUser updates a user on a target node.
func (u *User) PutNodeUser(
	ctx context.Context,
	request gen.PutNodeUserRequestObject,
) (gen.PutNodeUserResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeUser400JSONResponse{Error: &errMsg}, nil
	}

	opts := userProv.UpdateUserOpts{}
	if request.Body.Shell != nil {
		opts.Shell = *request.Body.Shell
	}
	if request.Body.Home != nil {
		opts.Home = *request.Body.Home
	}
	if request.Body.Groups != nil {
		opts.Groups = *request.Body.Groups
	}
	if request.Body.Lock != nil {
		opts.Lock = request.Body.Lock
	}

	hostname := request.Hostname
	name := request.Name

	u.logger.Debug("user update",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := map[string]interface{}{
		"name": name,
		"opts": opts,
	}

	if job.IsBroadcastTarget(hostname) {
		return u.putNodeUserUpdateBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := u.JobClient.Modify(
		ctx,
		hostname,
		"user",
		job.OperationUserUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.PutNodeUser404JSONResponse{Error: &errMsg}, nil
		}
		return gen.PutNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PutNodeUser200JSONResponse{
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

	return gen.PutNodeUser200JSONResponse{
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

// putNodeUserUpdateBroadcast handles broadcast targets for user update.
func (u *User) putNodeUserUpdateBroadcast(
	ctx context.Context,
	target string,
	data map[string]interface{},
) (gen.PutNodeUserResponseObject, error) {
	jobID, responses, err := u.JobClient.ModifyBroadcast(
		ctx,
		target,
		"user",
		job.OperationUserUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeUser500JSONResponse{Error: &errMsg}, nil
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

	return gen.PutNodeUser200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
