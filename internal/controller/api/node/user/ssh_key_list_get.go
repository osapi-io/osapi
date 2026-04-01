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
)

// GetNodeUserSSHKey lists SSH authorized keys for a user on a target node.
func (u *User) GetNodeUserSSHKey(
	ctx context.Context,
	request gen.GetNodeUserSSHKeyRequestObject,
) (gen.GetNodeUserSSHKeyResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	username := request.Name

	u.logger.Debug("ssh key list",
		slog.String("target", hostname),
		slog.String("username", username),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := map[string]string{
		"username": username,
	}

	if job.IsBroadcastTarget(hostname) {
		return u.getNodeUserSSHKeyBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := u.JobClient.Query(
		ctx,
		hostname,
		"user",
		job.OperationSSHKeyList,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeUserSSHKey200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.SSHKeyEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.SSHKeyEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := sshKeyInfoListFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUserSSHKey200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeUserSSHKeyBroadcast handles broadcast targets for SSH key list.
func (u *User) getNodeUserSSHKeyBroadcast(
	ctx context.Context,
	target string,
	data map[string]string,
) (gen.GetNodeUserSSHKeyResponseObject, error) {
	jobID, responses, err := u.JobClient.QueryBroadcast(
		ctx,
		target,
		"user",
		job.OperationSSHKeyList,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUserSSHKey500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.SSHKeyEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			allResults = append(allResults, gen.SSHKeyEntry{
				Hostname: host,
				Status:   gen.SSHKeyEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			allResults = append(allResults, gen.SSHKeyEntry{
				Hostname: host,
				Status:   gen.SSHKeyEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, sshKeyInfoListFromResponse(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUserSSHKey200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// sshKeyInfoListFromResponse converts a job response to gen SSHKeyEntry slice.
func sshKeyInfoListFromResponse(
	resp *job.Response,
) []gen.SSHKeyEntry {
	var keys []userProv.SSHKey
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &keys)
	}

	hostname := resp.Hostname

	keyInfos := make([]gen.SSHKeyInfo, 0, len(keys))
	for _, k := range keys {
		keyInfos = append(keyInfos, sshKeyInfoToGen(k))
	}

	return []gen.SSHKeyEntry{
		{
			Hostname: hostname,
			Status:   gen.SSHKeyEntryStatusOk,
			Keys:     &keyInfos,
		},
	}
}

// sshKeyInfoToGen converts a provider SSHKey to a gen SSHKeyInfo.
func sshKeyInfoToGen(
	k userProv.SSHKey,
) gen.SSHKeyInfo {
	keyType := k.Type
	fingerprint := k.Fingerprint

	info := gen.SSHKeyInfo{
		Type:        &keyType,
		Fingerprint: &fingerprint,
	}

	if k.Comment != "" {
		comment := k.Comment
		info.Comment = &comment
	}

	return info
}
