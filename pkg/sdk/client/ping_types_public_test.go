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

package client_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

type PingTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *PingTypesPublicTestSuite) TestPingCollectionFromGen() {
	tests := []struct {
		name         string
		input        *gen.PingCollectionResponse
		validateFunc func(client.Collection[client.PingResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.PingCollectionResponse {
				packetsSent := 5
				packetsReceived := 5
				packetLoss := 0.0
				minRtt := "1.234ms"
				avgRtt := "2.345ms"
				maxRtt := "3.456ms"
				changed := false

				return &gen.PingCollectionResponse{
					Results: []gen.PingResponse{
						{
							Hostname:        "web-01",
							Changed:         &changed,
							PacketsSent:     &packetsSent,
							PacketsReceived: &packetsReceived,
							PacketLoss:      &packetLoss,
							MinRtt:          &minRtt,
							AvgRtt:          &avgRtt,
							MaxRtt:          &maxRtt,
						},
					},
				}
			}(),
			validateFunc: func(c client.Collection[client.PingResult]) {
				suite.Require().Len(c.Results, 1)

				pr := c.Results[0]
				suite.Equal("web-01", pr.Hostname)
				suite.Equal(5, pr.PacketsSent)
				suite.Equal(5, pr.PacketsReceived)
				suite.InDelta(0.0, pr.PacketLoss, 0.001)
				suite.Equal("1.234ms", pr.MinRtt)
				suite.Equal("2.345ms", pr.AvgRtt)
				suite.Equal("3.456ms", pr.MaxRtt)
				suite.Empty(pr.Error)
				suite.False(pr.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportPingCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestPingTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PingTypesPublicTestSuite))
}
