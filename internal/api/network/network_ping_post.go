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
	"net/http"
	"time"

	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/validation"
)

// pingMultiResponse wraps multiple ping results for _all broadcast.
type pingMultiResponse struct {
	Results []gen.PingResponse `json:"results"`
}

func (r pingMultiResponse) VisitPostNetworkPingResponse(
	w http.ResponseWriter,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	return json.NewEncoder(w).Encode(r)
}

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

	if hostname == job.BroadcastHost {
		return n.postNetworkPingAll(ctx, request.Body.Address)
	}

	pingResult, err := n.JobClient.QueryNetworkPing(ctx, hostname, request.Body.Address)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return buildPingResponse(pingResult), nil
}

// postNetworkPingAll handles _all broadcast for ping.
func (n Network) postNetworkPingAll(
	ctx context.Context,
	address string,
) (gen.PostNetworkPingResponseObject, error) {
	results, err := n.JobClient.QueryNetworkPingAll(ctx, address)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.PingResponse
	for _, r := range results {
		responses = append(responses, gen.PingResponse(buildPingResponse(r)))
	}

	return pingMultiResponse{Results: responses}, nil
}

// buildPingResponse converts a ping.Result to the API response.
func buildPingResponse(
	r *ping.Result,
) gen.PostNetworkPing200JSONResponse {
	return gen.PostNetworkPing200JSONResponse{
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
