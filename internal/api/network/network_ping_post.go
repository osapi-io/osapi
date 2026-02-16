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
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/retr0h/osapi/internal/api/network/gen"
)

// PostNetworkPing post the network ping API endpoint.
func (n Network) PostNetworkPing(
	ctx context.Context,
	request gen.PostNetworkPingRequestObject,
) (gen.PostNetworkPingResponseObject, error) {
	validate := validator.New()
	if err := validate.Struct(request.Body); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errMsg := validationErrors.Error()
		return gen.PostNetworkPing400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	pingResult, err := n.JobClient.QueryNetworkPingAny(ctx, request.Body.Address)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNetworkPing500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.PostNetworkPing200JSONResponse{
		AvgRtt:          durationToString(&pingResult.AvgRTT),
		MaxRtt:          durationToString(&pingResult.MaxRTT),
		MinRtt:          durationToString(&pingResult.MinRTT),
		PacketLoss:      &pingResult.PacketLoss,
		PacketsReceived: &pingResult.PacketsReceived,
		PacketsSent:     &pingResult.PacketsSent,
	}, nil
}

// durationToString convert *time.Duration to *string.
func durationToString(d *time.Duration) *string {
	if d == nil {
		return nil
	}
	str := fmt.Sprintf("%v", *d)
	return &str
}
