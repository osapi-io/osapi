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

package client

import (
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// InterfaceInfo represents a network interface's configuration.
type InterfaceInfo struct {
	Name       string   `json:"name,omitempty"`
	DHCP4      bool     `json:"dhcp4"`
	DHCP6      bool     `json:"dhcp6"`
	Addresses  []string `json:"addresses,omitempty"`
	Gateway4   string   `json:"gateway4,omitempty"`
	Gateway6   string   `json:"gateway6,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	MACAddress string   `json:"mac_address,omitempty"`
	WakeOnLAN  bool     `json:"wakeonlan"`
	Managed    bool     `json:"managed"`
	State      string   `json:"state,omitempty"`
}

// InterfaceListResult represents an interface list result from a single agent.
type InterfaceListResult struct {
	Hostname   string          `json:"hostname"`
	Status     string          `json:"status"`
	Error      string          `json:"error,omitempty"`
	Interfaces []InterfaceInfo `json:"interfaces,omitempty"`
}

// InterfaceGetResult represents an interface get result from a single agent.
type InterfaceGetResult struct {
	Hostname  string         `json:"hostname"`
	Status    string         `json:"status"`
	Error     string         `json:"error,omitempty"`
	Interface *InterfaceInfo `json:"interface,omitempty"`
}

// InterfaceMutationResult represents the result of an interface
// create, update, or delete.
type InterfaceMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// InterfaceConfigOpts contains options for creating or updating
// a network interface.
type InterfaceConfigOpts struct {
	// DHCP4 enables or disables DHCPv4.
	DHCP4 *bool
	// DHCP6 enables or disables DHCPv6.
	DHCP6 *bool
	// Addresses is the list of IP addresses in CIDR notation.
	Addresses []string
	// Gateway4 is the IPv4 gateway address.
	Gateway4 string
	// Gateway6 is the IPv6 gateway address.
	Gateway6 string
	// MTU is the maximum transmission unit.
	MTU *int
	// MACAddress is the hardware MAC address.
	MACAddress string
	// WakeOnLAN enables or disables Wake-on-LAN.
	WakeOnLAN *bool
}

// interfaceInfoFromGen converts a gen.InterfaceInfo to an InterfaceInfo.
func interfaceInfoFromGen(
	g gen.InterfaceInfo,
) InterfaceInfo {
	info := InterfaceInfo{
		Name:       derefString(g.Name),
		DHCP4:      derefBool(g.Dhcp4),
		DHCP6:      derefBool(g.Dhcp6),
		Gateway4:   derefString(g.Gateway4),
		Gateway6:   derefString(g.Gateway6),
		MTU:        derefInt(g.Mtu),
		MACAddress: derefString(g.MacAddress),
		WakeOnLAN:  derefBool(g.Wakeonlan),
		Managed:    derefBool(g.Managed),
		State:      derefString(g.State),
	}

	if g.Addresses != nil {
		info.Addresses = *g.Addresses
	}

	return info
}

// interfaceListCollectionFromGen converts a gen.InterfaceListResponse
// to a Collection[InterfaceListResult].
func interfaceListCollectionFromGen(
	g *gen.InterfaceListResponse,
) Collection[InterfaceListResult] {
	results := make([]InterfaceListResult, 0, len(g.Results))
	for _, r := range g.Results {
		entry := InterfaceListResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Interfaces != nil {
			ifaces := make([]InterfaceInfo, 0, len(*r.Interfaces))
			for _, iface := range *r.Interfaces {
				ifaces = append(ifaces, interfaceInfoFromGen(iface))
			}

			entry.Interfaces = ifaces
		}

		results = append(results, entry)
	}

	return Collection[InterfaceListResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// interfaceGetCollectionFromGen converts a gen.InterfaceGetResponse
// to a Collection[InterfaceGetResult].
func interfaceGetCollectionFromGen(
	g *gen.InterfaceGetResponse,
) Collection[InterfaceGetResult] {
	results := make([]InterfaceGetResult, 0, len(g.Results))
	for _, r := range g.Results {
		entry := InterfaceGetResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Interface != nil {
			info := interfaceInfoFromGen(*r.Interface)
			entry.Interface = &info
		}

		results = append(results, entry)
	}

	return Collection[InterfaceGetResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// interfaceMutationCollectionFromCreate converts a
// gen.InterfaceMutationResponse to a Collection[InterfaceMutationResult].
func interfaceMutationCollectionFromCreate(
	g *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromGen(g)
}

// interfaceMutationCollectionFromUpdate converts a
// gen.InterfaceMutationResponse to a Collection[InterfaceMutationResult].
func interfaceMutationCollectionFromUpdate(
	g *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromGen(g)
}

// interfaceMutationCollectionFromDelete converts a
// gen.InterfaceMutationResponse to a Collection[InterfaceMutationResult].
func interfaceMutationCollectionFromDelete(
	g *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	return interfaceMutationCollectionFromGen(g)
}

// interfaceMutationCollectionFromGen converts a gen.InterfaceMutationResponse
// to a Collection[InterfaceMutationResult].
func interfaceMutationCollectionFromGen(
	g *gen.InterfaceMutationResponse,
) Collection[InterfaceMutationResult] {
	results := make([]InterfaceMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, InterfaceMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[InterfaceMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}
