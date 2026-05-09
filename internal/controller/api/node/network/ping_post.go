// Copyright (c) 2024 John Dewey

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

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeNetworkPing post the node network ping API endpoint.
func (s *Network) PostNodeNetworkPing(
	ctx context.Context,
	request gen.PostNodeNetworkPingRequestObject,
) (gen.PostNodeNetworkPingResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeNetworkPing400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeNetworkPing400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"ping",
		slog.String("address", request.Body.Address),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeNetworkPingBroadcast(ctx, hostname, request.Body.Address)
	}

	data := map[string]any{"address": request.Body.Address}
	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkPingDo,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.PostNodeNetworkPing200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.PingResponse{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.PingResponseStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var pingResult ping.Result
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &pingResult)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeNetworkPing200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.PingResponse{
			buildPingResponse(rawResp.Hostname, &pingResult),
		},
	}, nil
}

// postNodeNetworkPingBroadcast handles broadcast targets (_all or label) for ping.
func (s *Network) postNodeNetworkPingBroadcast(
	ctx context.Context,
	target string,
	address string,
) (gen.PostNodeNetworkPingResponseObject, error) {
	data := map[string]any{"address": address}
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkPingDo,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := make([]gen.PingResponse, 0, len(responses))
	for host, resp := range responses {
		item := gen.PingResponse{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.PingResponseStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.PingResponseStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.PingResponseStatusOk
			var pingResult ping.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &pingResult)
			}
			built := buildPingResponse(host, &pingResult)
			item.AvgRtt = built.AvgRtt
			item.MaxRtt = built.MaxRtt
			item.MinRtt = built.MinRtt
			item.PacketLoss = built.PacketLoss
			item.PacketsReceived = built.PacketsReceived
			item.PacketsSent = built.PacketsSent
			item.Changed = built.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeNetworkPing200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// buildPingResponse converts a ping.Result to the API response.
func buildPingResponse(
	hostname string,
	r *ping.Result,
) gen.PingResponse {
	changed := false

	return gen.PingResponse{
		Hostname:        hostname,
		Status:          gen.PingResponseStatusOk,
		AvgRtt:          durationToString(&r.AvgRTT),
		MaxRtt:          durationToString(&r.MaxRTT),
		MinRtt:          durationToString(&r.MinRTT),
		PacketLoss:      &r.PacketLoss,
		PacketsReceived: &r.PacketsReceived,
		PacketsSent:     &r.PacketsSent,
		Changed:         &changed,
	}
}

// durationToString converts *time.Duration to a human-readable *string
// with consistent precision (e.g., "22.63ms").
func durationToString(
	d *time.Duration,
) *string {
	if d == nil {
		return nil
	}
	ms := float64(*d) / float64(time.Millisecond)
	str := fmt.Sprintf("%.2fms", ms)
	return &str
}
