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

// GetNodeGroupByName gets a single group by name on a target node.
func (u *User) GetNodeGroupByName(
	ctx context.Context,
	request gen.GetNodeGroupByNameRequestObject,
) (gen.GetNodeGroupByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeGroupByName500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	u.logger.Debug("group get",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return u.getNodeGroupByNameBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := u.JobClient.Query(
		ctx,
		hostname,
		"group",
		job.OperationGroupGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.GetNodeGroupByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeGroupByName500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeGroupByName200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.GroupEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.GroupEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var entry userProv.Group
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entry)
	}

	jobUUID := uuid.MustParse(jobID)
	groupName := entry.Name
	gid := entry.GID
	agentHostname := resp.Hostname

	info := gen.GroupInfo{
		Name: &groupName,
		Gid:  &gid,
	}
	if len(entry.Members) > 0 {
		members := entry.Members
		info.Members = &members
	}

	return gen.GetNodeGroupByName200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.GroupEntry{
			{
				Hostname: agentHostname,
				Status:   gen.GroupEntryStatusOk,
				Groups:   &[]gen.GroupInfo{info},
			},
		},
	}, nil
}

// getNodeGroupByNameBroadcast handles broadcast targets for group get.
func (u *User) getNodeGroupByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodeGroupByNameResponseObject, error) {
	jobID, responses, err := u.JobClient.QueryBroadcast(
		ctx,
		target,
		"group",
		job.OperationGroupGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeGroupByName500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.GroupEntry, 0)
	for host, resp := range responses {
		item := gen.GroupEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.GroupEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.GroupEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.GroupEntryStatusOk
			var entry userProv.Group
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			groupName := entry.Name
			gid := entry.GID
			info := gen.GroupInfo{
				Name: &groupName,
				Gid:  &gid,
			}
			if len(entry.Members) > 0 {
				members := entry.Members
				info.Members = &members
			}
			item.Groups = &[]gen.GroupInfo{info}
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeGroupByName200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}
