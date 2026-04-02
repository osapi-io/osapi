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

// GetNodeGroup lists all groups on a target node.
func (u *User) GetNodeGroup(
	ctx context.Context,
	request gen.GetNodeGroupRequestObject,
) (gen.GetNodeGroupResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeGroup400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	u.logger.Debug("group list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return u.getNodeGroupBroadcast(ctx, hostname)
	}

	jobID, resp, err := u.JobClient.Query(ctx, hostname, "group", job.OperationGroupList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeGroup500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeGroup200JSONResponse{
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

	results := responseToGroupEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeGroup200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeGroupBroadcast handles broadcast targets for group list.
func (u *User) getNodeGroupBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeGroupResponseObject, error) {
	jobID, responses, err := u.JobClient.QueryBroadcast(
		ctx,
		target,
		"group",
		job.OperationGroupList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeGroup500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.GroupEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.GroupEntry{
				Hostname: h,
				Status:   gen.GroupEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.GroupEntry{
				Hostname: h,
				Status:   gen.GroupEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToGroupEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeGroup200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToGroupEntries converts a job response to gen GroupEntry slice.
func responseToGroupEntries(
	resp *job.Response,
) []gen.GroupEntry {
	var groups []userProv.Group
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &groups)
	}

	hostname := resp.Hostname

	groupInfos := make([]gen.GroupInfo, 0, len(groups))
	for _, g := range groups {
		name := g.Name
		gid := g.GID

		info := gen.GroupInfo{
			Name: &name,
			Gid:  &gid,
		}
		if len(g.Members) > 0 {
			members := g.Members
			info.Members = &members
		}

		groupInfos = append(groupInfos, info)
	}

	return []gen.GroupEntry{
		{
			Hostname: hostname,
			Status:   gen.GroupEntryStatusOk,
			Groups:   &groupInfos,
		},
	}
}
