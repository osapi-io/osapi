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

package route

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// spec is a route of a Config, read from its string.
type spec struct {
	to     string
	via    string
	metric int
}

// parse reads a route as ip route writes it: <destination> via <gateway>,
// with metric <n> after it when given. The destination is in CIDR notation,
// and the gateway is of the same family.
func parse(
	s string,
) (spec, error) {
	f := strings.Fields(s)
	if (len(f) != 3 && len(f) != 5) || f[1] != "via" || (len(f) == 5 && f[3] != "metric") {
		return spec{}, fmt.Errorf("route %q is not <destination> via <gateway> [metric <n>]", s)
	}
	_, dest, err := net.ParseCIDR(f[0])
	if err != nil {
		return spec{}, fmt.Errorf("route %q: the destination is not in CIDR notation", s)
	}
	gw := net.ParseIP(f[2])
	if gw == nil || (gw.To4() == nil) != (dest.IP.To4() == nil) {
		return spec{}, fmt.Errorf(
			"route %q: the gateway is not an address of the destination's family",
			s,
		)
	}
	r := spec{to: f[0], via: f[2]}
	if len(f) == 5 {
		n, err := strconv.ParseUint(f[4], 10, 32)
		if err != nil {
			return spec{}, fmt.Errorf(
				"route %q: the metric is not a number from 0 to 4294967295",
				s,
			)
		}
		r.metric = int(n)
	}

	return r, nil
}

// valid reports whether s is a route that parse reads.
func valid(
	s string,
) bool {
	_, err := parse(s)

	return err == nil
}

// isDefault reports whether the route is a default route, which osapi does
// not write: a wrong one cuts the host off.
func (r spec) isDefault() bool {
	_, n, _ := net.ParseCIDR(r.to)
	ones, _ := n.Mask.Size()

	return ones == 0
}
