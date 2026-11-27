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
)

// spec is one route, as netplan writes it.
type spec struct {
	To     string `yaml:"to"`
	Via    string `yaml:"via"`
	Metric int    `yaml:"metric,omitempty"`
}

func toSpec(
	r Route,
) spec {
	return spec{To: canon(r.Destination), Via: r.Via, Metric: r.Metric}
}

// canon is the network of a destination, in one spelling: 10.0.0.5/8 is
// 10.0.0.0/8, and an IPv6 network is in lower case and compressed. Thus two
// spellings of one route are the same route. A destination that does not
// parse stays as it is, for check and the validate tag to refuse.
func canon(
	destination string,
) string {
	_, n, err := net.ParseCIDR(destination)
	if err != nil {
		return destination
	}

	return n.String()
}

// check refuses a default route, which osapi does not write: a wrong one cuts
// the host off, and the interface behavior sets the gateway. It refuses a
// gateway of the other family, which the validate tags of the two fields
// cannot see. The tags have checked each one alone.
func check(
	r spec,
) error {
	_, dest, _ := net.ParseCIDR(r.To)
	if ones, _ := dest.Mask.Size(); ones == 0 {
		return fmt.Errorf("%s is a default route, which osapi does not write: "+
			"give the interface a gateway with osapi client node interface", r.To)
	}
	if (net.ParseIP(r.Via).To4() == nil) != (dest.IP.To4() == nil) {
		return fmt.Errorf("the gateway %s is not an address of the family of %s", r.Via, r.To)
	}

	return nil
}
