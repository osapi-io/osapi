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

package netplan

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/retr0h/osapi/internal/exec"
)

// Status holds the parsed output of netplan status --format json.
// Keys are interface names, values are the interface status.
type Status map[string]InterfaceStatus

// InterfaceStatus represents one interface from netplan status.
type InterfaceStatus struct {
	Index        int                      `json:"index"`
	AdminState   string                   `json:"adminstate"`
	OperState    string                   `json:"operstate"`
	Type         string                   `json:"type"`
	Backend      string                   `json:"backend"`
	ID           string                   `json:"id"`
	MACAddress   string                   `json:"macaddress"`
	Vendor       string                   `json:"vendor"`
	Addresses    []map[string]AddressInfo `json:"addresses"`
	DNSAddresses []string                 `json:"dns_addresses"`
	Routes       []RouteStatus            `json:"routes"`
	Interfaces   []string                 `json:"interfaces"`
	Bridge       string                   `json:"bridge"`
	TunnelMode   string                   `json:"tunnel_mode"`
}

// AddressInfo holds prefix length and flags for an address.
type AddressInfo struct {
	Prefix int      `json:"prefix"`
	Flags  []string `json:"flags"`
}

// RouteStatus represents a single route from netplan status.
type RouteStatus struct {
	To       string `json:"to"`
	Via      string `json:"via"`
	From     string `json:"from"`
	Family   int    `json:"family"`
	Metric   int    `json:"metric"`
	Type     string `json:"type"`
	Scope    string `json:"scope"`
	Protocol string `json:"protocol"`
	Table    string `json:"table"`
}

// GetStatus runs "netplan status --format json" and parses the output.
func GetStatus(
	execManager exec.Manager,
) (Status, error) {
	output, err := execManager.RunCmd(
		"netplan",
		[]string{"status", "--format", "json"},
	)
	if err != nil {
		return nil, fmt.Errorf("netplan status: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil, fmt.Errorf("parse netplan status: %w", err)
	}

	result := make(Status)

	for key, val := range raw {
		if key == "netplan-global-state" {
			continue
		}

		var iface InterfaceStatus
		if err := json.Unmarshal(val, &iface); err != nil {
			continue
		}

		result[key] = iface
	}

	return result, nil
}

// IPv4 returns the first non-link-local IPv4 address for this interface.
func (s InterfaceStatus) IPv4() string {
	for _, addrMap := range s.Addresses {
		for addr := range addrMap {
			ip := net.ParseIP(addr)
			if ip == nil {
				continue
			}

			if ip.To4() != nil && !ip.IsLinkLocalUnicast() {
				return addr
			}
		}
	}

	return ""
}

// IPv6 returns the first non-link-local IPv6 address for this interface.
func (s InterfaceStatus) IPv6() string {
	for _, addrMap := range s.Addresses {
		for addr := range addrMap {
			ip := net.ParseIP(addr)
			if ip == nil {
				continue
			}

			if ip.To4() == nil && !ip.IsLinkLocalUnicast() {
				return addr
			}
		}
	}

	return ""
}

// IsDHCP returns true if any address has a "dhcp" flag.
func (s InterfaceStatus) IsDHCP() bool {
	for _, addrMap := range s.Addresses {
		for _, info := range addrMap {
			for _, flag := range info.Flags {
				if strings.EqualFold(flag, "dhcp") {
					return true
				}
			}
		}
	}

	return false
}

// HasDefaultRoute returns true if any route has To == "default".
func (s InterfaceStatus) HasDefaultRoute() bool {
	for _, r := range s.Routes {
		if r.To == "default" {
			return true
		}
	}

	return false
}

// AddressFamily returns "inet" for IPv4, "inet6" for IPv6, or "inet" if
// the first non-link-local address cannot be determined.
func (s InterfaceStatus) AddressFamily() string {
	for _, addrMap := range s.Addresses {
		for addr := range addrMap {
			ip := net.ParseIP(addr)
			if ip == nil || ip.IsLinkLocalUnicast() {
				continue
			}

			if ip.To4() != nil {
				return "inet"
			}

			return "inet6"
		}
	}

	return "inet"
}

// SectionForInterface returns the Netplan YAML section name for an
// interface by querying netplan status. Falls back to "ethernets" if
// the interface type cannot be determined.
func SectionForInterface(
	execManager exec.Manager,
	interfaceName string,
) string {
	status, err := GetStatus(execManager)
	if err != nil {
		return "ethernets"
	}

	iface, ok := status[interfaceName]
	if !ok {
		return "ethernets"
	}

	return SectionForType(iface.Type)
}

// SectionForType maps a netplan interface type to its YAML section name.
func SectionForType(
	ifaceType string,
) string {
	switch ifaceType {
	case "wifi":
		return "wifis"
	case "bridge":
		return "bridges"
	case "bond":
		return "bonds"
	case "tunnel", "vxlan":
		return "tunnels"
	default:
		return "ethernets"
	}
}
