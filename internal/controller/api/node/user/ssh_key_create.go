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

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/user/gen"
	"github.com/retr0h/osapi/internal/job"
	userProv "github.com/retr0h/osapi/internal/provider/node/user"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeUserSSHKey adds an SSH authorized key for a user on a target node.
func (u *User) PostNodeUserSSHKey(
	ctx context.Context,
	request gen.PostNodeUserSSHKeyRequestObject,
) (gen.PostNodeUserSSHKeyResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeUserSSHKey400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeUserSSHKey400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	username := request.Name

	u.logger.Debug("ssh key add",
		slog.String("target", hostname),
		slog.String("username", username),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := map[string]string{
		"username": username,
		"raw_line": request.Body.Key,
	}

	if job.IsBroadcastTarget(hostname) {
		return u.postNodeUserSSHKeyBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := u.JobClient.Modify(
		ctx,
		hostname,
		"user",
		job.OperationSSHKeyAdd,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeUserSSHKey200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.SSHKeyMutationEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.SSHKeyMutationEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result userProv.SSHKeyResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	return gen.PostNodeUserSSHKey200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.SSHKeyMutationEntry{
			{
				Hostname: agentHostname,
				Status:   gen.SSHKeyMutationEntryStatusOk,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodeUserSSHKeyBroadcast handles broadcast targets for SSH key add.
func (u *User) postNodeUserSSHKeyBroadcast(
	ctx context.Context,
	target string,
	data map[string]string,
) (gen.PostNodeUserSSHKeyResponseObject, error) {
	jobID, responses, err := u.JobClient.ModifyBroadcast(
		ctx,
		target,
		"user",
		job.OperationSSHKeyAdd,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.SSHKeyMutationEntry
	for host, resp := range responses {
		item := gen.SSHKeyMutationEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.SSHKeyMutationEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.SSHKeyMutationEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.SSHKeyMutationEntryStatusOk
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodeUserSSHKey200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
