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

package ping

import (
	"context"
	"fmt"
	"time"
)

// replyTimeout is how long a ping waits for its replies. It is below
// DefaultTimeout, so the pinger stops first and reports its statistics.
const replyTimeout = 4 * time.Second

// DefaultTimeout is how long Do waits for the whole ping.
const DefaultTimeout = 5 * time.Second

// Do pings the given host and returns the ping statistics or an error.
// It returns a Result struct, and an error if something goes wrong.
//
// On Linux, it attempts an "unprivileged" ping via UDP by default.
// To enable unprivileged pings, run the following command:
//
//	sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
//
// Alternatively, call pinger.SetPrivileged(true) to use raw sockets.
// After doing so, you may either:
//  1. Run the binary as root, or
//  2. Set capabilities on the binary to allow raw socket usage:
//     sudo setcap cap_net_raw=+ep /path/to/your/compiled/binary
func (u *Debian) Do(
	job context.Context,
	address string,
) (*Result, error) {
	pinger, err := u.NewPingerFn(address)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pinger: %w", err)
	}

	pinger.SetCount(3)
	pinger.SetPrivileged(false)
	// The pinger stops at its own deadline, before the outer timeout, and
	// reports a reply that did not come back as loss. Without it, one lost
	// reply kept Run waiting until the outer timeout failed the whole ping.
	pinger.SetTimeout(replyTimeout)

	// Create a context with a timeout for the entire ping operation
	timeout := u.Timeout
	// The timeout is under the job's context, so a job that ends stops it.
	ctx, cancel := context.WithTimeout(job, timeout)
	defer cancel()

	// Run the ping in a separate goroutine to avoid blocking
	// Buffered, so the goroutine can finish after a timeout. Do returns as
	// soon as the context expires and stops receiving; on an unbuffered
	// channel the send would then block forever and leak the goroutine on
	// every timed-out ping.
	resultChan := make(chan *Result, 1)
	errorChan := make(chan error, 1)

	go func() {
		err = pinger.Run() // This blocks until pinging is finished
		if err != nil {
			errorChan <- fmt.Errorf("failed to run pinger: %w", err)
			return
		}

		stats := pinger.Statistics()
		result := &Result{
			PacketsSent:     stats.PacketsSent,
			PacketsReceived: stats.PacketsRecv,
			PacketLoss:      stats.PacketLoss,
			MinRTT:          stats.MinRtt,
			AvgRTT:          stats.AvgRtt,
			MaxRTT:          stats.MaxRtt,
		}

		resultChan <- result
	}()

	// Use a select statement to enforce the timeout
	select {
	case <-ctx.Done():
		if err := job.Err(); err != nil {
			return nil, fmt.Errorf("ping stopped, the job ended: %w", err)
		}

		return nil, fmt.Errorf("ping operation timed out after %s", timeout)
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}
