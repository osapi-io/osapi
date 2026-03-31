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

// GetNodeUserByName gets a single user by name on a target node.
func (u *User) GetNodeUserByName(
	ctx context.Context,
	request gen.GetNodeUserByNameRequestObject,
) (gen.GetNodeUserByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeUserByName500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	u.logger.Debug("user get",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return u.getNodeUserByNameBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := u.JobClient.Query(
		ctx,
		hostname,
		"user",
		job.OperationUserGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.GetNodeUserByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeUserByName500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeUserByName200JSONResponse{
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

	var entry userProv.User
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entry)
	}

	jobUUID := uuid.MustParse(jobID)
	userName := entry.Name
	uid := entry.UID
	gid := entry.GID
	home := entry.Home
	shell := entry.Shell
	locked := entry.Locked
	agentHostname := resp.Hostname

	info := gen.UserInfo{
		Name:   &userName,
		Uid:    &uid,
		Gid:    &gid,
		Home:   &home,
		Shell:  &shell,
		Locked: &locked,
	}
	if len(entry.Groups) > 0 {
		groups := entry.Groups
		info.Groups = &groups
	}

	return gen.GetNodeUserByName200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.UserEntry{
			{
				Hostname: agentHostname,
				Status:   gen.UserEntryStatusOk,
				Users:    &[]gen.UserInfo{info},
			},
		},
	}, nil
}

// getNodeUserByNameBroadcast handles broadcast targets for user get.
func (u *User) getNodeUserByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodeUserByNameResponseObject, error) {
	jobID, responses, err := u.JobClient.QueryBroadcast(
		ctx,
		target,
		"user",
		job.OperationUserGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUserByName500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.UserEntry, 0)
	for host, resp := range responses {
		item := gen.UserEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.UserEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.UserEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.UserEntryStatusOk
			var entry userProv.User
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			userName := entry.Name
			uid := entry.UID
			gid := entry.GID
			home := entry.Home
			shell := entry.Shell
			locked := entry.Locked
			info := gen.UserInfo{
				Name:   &userName,
				Uid:    &uid,
				Gid:    &gid,
				Home:   &home,
				Shell:  &shell,
				Locked: &locked,
			}
			if len(entry.Groups) > 0 {
				groups := entry.Groups
				info.Groups = &groups
			}
			item.Users = &[]gen.UserInfo{info}
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUserByName200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}
