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
	"fmt"
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNetworkPing post the network ping API endpoint.
func (n Network) PostNetworkPing(
	ctx context.Context,
	request gen.PostNetworkPingRequestObject,
) (gen.PostNetworkPingResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNetworkPing400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.PostNetworkPing400JSONResponse{Error: &errMsg}, nil
		}
	}

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	n.logger.Debug("ping",
		slog.String("address", request.Body.Address),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return n.postNetworkPingBroadcast(ctx, hostname, request.Body.Address)
	}

	pingResult, workerHostname, err := n.JobClient.QueryNetworkPing(
		ctx,
		hostname,
		request.Body.Address,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PostNetworkPing200JSONResponse{
		Results: []gen.PingResponse{
			buildPingResponse(workerHostname, pingResult),
		},
	}, nil
}

// postNetworkPingBroadcast handles broadcast targets (_all or label) for ping.
func (n Network) postNetworkPingBroadcast(
	ctx context.Context,
	target string,
	address string,
) (gen.PostNetworkPingResponseObject, error) {
	results, err := n.JobClient.QueryNetworkPingBroadcast(ctx, target, address)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.PingResponse
	for host, r := range results {
		responses = append(responses, buildPingResponse(host, r))
	}

	return gen.PostNetworkPing200JSONResponse{
		Results: responses,
	}, nil
}

// buildPingResponse converts a ping.Result to the API response.
func buildPingResponse(
	hostname string,
	r *ping.Result,
) gen.PingResponse {
	return gen.PingResponse{
		Hostname:        hostname,
		AvgRtt:          durationToString(&r.AvgRTT),
		MaxRtt:          durationToString(&r.MaxRTT),
		MinRtt:          durationToString(&r.MinRTT),
		PacketLoss:      &r.PacketLoss,
		PacketsReceived: &r.PacketsReceived,
		PacketsSent:     &r.PacketsSent,
	}
}

// durationToString convert *time.Duration to *string.
func durationToString(
	d *time.Duration,
) *string {
	if d == nil {
		return nil
	}
	str := fmt.Sprintf("%v", *d)
	return &str
}
