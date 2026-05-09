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

// GetNodeUser lists all users on a target node.
func (u *User) GetNodeUser(
	ctx context.Context,
	request gen.GetNodeUserRequestObject,
) (gen.GetNodeUserResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeUser400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	u.logger.Debug(
		"user list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return u.getNodeUserBroadcast(ctx, hostname)
	}

	jobID, resp, err := u.JobClient.Query(ctx, hostname, "user", job.OperationUserList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeUser200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.UserEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.UserEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToUserEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUser200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeUserBroadcast handles broadcast targets for user list.
func (u *User) getNodeUserBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeUserResponseObject, error) {
	jobID, responses, err := u.JobClient.QueryBroadcast(
		ctx,
		target,
		"user",
		job.OperationUserList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUser500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.UserEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.UserEntry{
				Hostname: h,
				Status:   gen.UserEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.UserEntry{
				Hostname: h,
				Status:   gen.UserEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToUserEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUser200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToUserEntries converts a job response to gen UserEntry slice.
func responseToUserEntries(
	resp *job.Response,
) []gen.UserEntry {
	var users []userProv.User
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &users)
	}

	hostname := resp.Hostname

	userInfos := make([]gen.UserInfo, 0, len(users))
	for _, u := range users {
		name := u.Name
		uid := u.UID
		gid := u.GID
		home := u.Home
		shell := u.Shell
		locked := u.Locked

		info := gen.UserInfo{
			Name:   &name,
			Uid:    &uid,
			Gid:    &gid,
			Home:   &home,
			Shell:  &shell,
			Locked: &locked,
		}
		if len(u.Groups) > 0 {
			groups := u.Groups
			info.Groups = &groups
		}

		userInfos = append(userInfos, info)
	}

	return []gen.UserEntry{
		{
			Hostname: hostname,
			Status:   gen.UserEntryStatusOk,
			Users:    &userInfos,
		},
	}
}
