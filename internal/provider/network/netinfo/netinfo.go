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

// Package netinfo provides network interface information.
package netinfo

import (
	"net"

	"github.com/retr0h/osapi/internal/job"
)

// Netinfo implements the Provider interface for network interface information.
type Netinfo struct {
	InterfacesFn func() ([]net.Interface, error)
}

// New factory to create a new Netinfo instance.
func New() *Netinfo {
	return &Netinfo{
		InterfacesFn: net.Interfaces,
	}
}

// GetInterfaces retrieves non-loopback, up network interfaces
// with name, IPv4, and MAC address.
func (n *Netinfo) GetInterfaces() ([]job.NetworkInterface, error) {
	ifaces, err := n.InterfacesFn()
	if err != nil {
		return nil, err
	}

	var result []job.NetworkInterface

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		ni := job.NetworkInterface{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}

				if ipNet.IP.To4() != nil && ni.IPv4 == "" {
					ni.IPv4 = ipNet.IP.String()
				} else if ipNet.IP.To4() == nil && ni.IPv6 == "" {
					ni.IPv6 = ipNet.IP.String()
				}
			}
		}

		switch {
		case ni.IPv4 != "" && ni.IPv6 != "":
			ni.Family = "dual"
		case ni.IPv6 != "":
			ni.Family = "inet6"
		case ni.IPv4 != "":
			ni.Family = "inet"
		}

		result = append(result, ni)
	}

	return result, nil
}
