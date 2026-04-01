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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/user/gen"
	"github.com/retr0h/osapi/internal/job"
)

// DeleteNodeUserSSHKey removes an SSH authorized key by fingerprint for a user
// on a target node.
func (u *User) DeleteNodeUserSSHKey(
	ctx context.Context,
	request gen.DeleteNodeUserSSHKeyRequestObject,
) (gen.DeleteNodeUserSSHKeyResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	username := request.Name
	fingerprint := request.Fingerprint

	u.logger.Debug("ssh key remove",
		slog.String("target", hostname),
		slog.String("username", username),
		slog.String("fingerprint", fingerprint),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := map[string]string{
		"username":    username,
		"fingerprint": fingerprint,
	}

	if job.IsBroadcastTarget(hostname) {
		return u.deleteNodeUserSSHKeyBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := u.JobClient.Modify(
		ctx,
		hostname,
		"user",
		job.OperationSSHKeyRemove,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.DeleteNodeUserSSHKey200JSONResponse{
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

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	return gen.DeleteNodeUserSSHKey200JSONResponse{
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

// deleteNodeUserSSHKeyBroadcast handles broadcast targets for SSH key remove.
func (u *User) deleteNodeUserSSHKeyBroadcast(
	ctx context.Context,
	target string,
	data map[string]string,
) (gen.DeleteNodeUserSSHKeyResponseObject, error) {
	jobID, responses, err := u.JobClient.ModifyBroadcast(
		ctx,
		target,
		"user",
		job.OperationSSHKeyRemove,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
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

	return gen.DeleteNodeUserSSHKey200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}
