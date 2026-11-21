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
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// InterfaceInfo is a network interface as its host reports it.
type InterfaceInfo struct {
	Name      string   `json:"name,omitempty"`
	State     string   `json:"state,omitempty"`
	MAC       string   `json:"mac,omitempty"`
	MTU       int      `json:"mtu,omitempty"`
	IPv4      string   `json:"ipv4,omitempty"`
	IPv6      string   `json:"ipv6,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
	DHCP4     bool     `json:"dhcp4"`
	Primary   bool     `json:"primary"`
}

// InterfaceListResult is the network interfaces of one host.
type InterfaceListResult struct {
	Hostname   string          `json:"hostname"`
	Status     string          `json:"status"`
	Error      string          `json:"error,omitempty"`
	Interfaces []InterfaceInfo `json:"interfaces,omitempty"`
}

// InterfaceGetResult is one network interface of one host.
type InterfaceGetResult struct {
	Hostname  string         `json:"hostname"`
	Status    string         `json:"status"`
	Error     string         `json:"error,omitempty"`
	Interface *InterfaceInfo `json:"interface,omitempty"`
}

// InterfaceMutationResult is an interface after a create, an update or a
// delete on one host.
type InterfaceMutationResult struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Name     string `json:"name,omitempty"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// InterfaceConfigOpts is the config osapi writes for an interface. dhcp4
// and dhcp6 are false unless set.
type InterfaceConfigOpts struct {
	// DHCP4 takes an IPv4 address from DHCP.
	DHCP4 *bool
	// DHCP6 takes an IPv6 address from DHCP.
	DHCP6 *bool
	// Addresses is the static addresses, in CIDR notation.
	Addresses []string
	// Gateway4 is the IPv4 default gateway.
	Gateway4 string
	// Gateway6 is the IPv6 default gateway.
	Gateway6 string
	// MTU is the MTU, from 68 to 9000.
	MTU *int
	// MACAddress is the MAC address to set.
	MACAddress string
	// WakeOnLAN turns on wake-on-LAN.
	WakeOnLAN *bool
}

// interfaceRequest is the request body of opts.
func interfaceRequest(
	opts InterfaceConfigOpts,
) gen.NetworkInterfaceUpdateRequest {
	body := gen.NetworkInterfaceUpdateRequest{
		Dhcp4:     opts.DHCP4,
		Dhcp6:     opts.DHCP6,
		Mtu:       opts.MTU,
		Wakeonlan: opts.WakeOnLAN,
	}
	if len(opts.Addresses) > 0 {
		body.Addresses = &opts.Addresses
	}
	if opts.Gateway4 != "" {
		body.Gateway4 = &opts.Gateway4
	}
	if opts.Gateway6 != "" {
		body.Gateway6 = &opts.Gateway6
	}
	if opts.MACAddress != "" {
		body.MacAddress = &opts.MACAddress
	}

	return body
}

// interfaceInfoFromGen is the interface of a row, or nil for a row with
// none, such as a host that failed.
func interfaceInfoFromGen(
	r gen.NetworkInterfaceEntry,
) *InterfaceInfo {
	if r.Name == nil {
		return nil
	}

	return &InterfaceInfo{
		Name:      *r.Name,
		State:     derefString(r.State),
		MAC:       derefString(r.Mac),
		MTU:       derefInt(r.Mtu),
		IPv4:      derefString(r.Ipv4),
		IPv6:      derefString(r.Ipv6),
		Addresses: derefStrings(r.Addresses),
		DHCP4:     derefBool(r.Dhcp4),
		Primary:   derefBool(r.Primary),
	}
}

// interfaceListFromGen groups the rows of a list, one per interface, into
// one result per host, in the order each host first answers.
func interfaceListFromGen(
	g *gen.NetworkInterfaceCollectionResponse,
) Collection[InterfaceListResult] {
	results := []InterfaceListResult{}
	index := map[string]int{}
	for _, r := range g.Results {
		i, ok := index[r.Hostname]
		if !ok {
			i = len(results)
			index[r.Hostname] = i
			results = append(results, InterfaceListResult{
				Hostname: r.Hostname,
				Status:   string(r.Status),
				Error:    derefString(r.Error),
			})
		}
		if info := interfaceInfoFromGen(r); info != nil {
			results[i].Interfaces = append(results[i].Interfaces, *info)
		}
	}

	return Collection[InterfaceListResult]{Results: results, JobID: jobIDFromGen(g.JobId)}
}

// interfaceGetFromGen is the rows of a get.
func interfaceGetFromGen(
	g *gen.NetworkInterfaceGetResponse,
) Collection[InterfaceGetResult] {
	results := make([]InterfaceGetResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, InterfaceGetResult{
			Hostname:  r.Hostname,
			Status:    string(r.Status),
			Error:     derefString(r.Error),
			Interface: interfaceInfoFromGen(r),
		})
	}

	return Collection[InterfaceGetResult]{Results: results, JobID: jobIDFromGen(g.JobId)}
}

// interfaceMutationsFromGen is the rows of a create, an update or a delete.
func interfaceMutationsFromGen(
	rows []gen.NetworkInterfaceResult,
	jobID *openapi_types.UUID,
) Collection[InterfaceMutationResult] {
	results := make([]InterfaceMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, InterfaceMutationResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Name:     derefString(r.Name),
			Changed:  derefBool(r.Changed),
			Error:    derefString(r.Error),
		})
	}

	return Collection[InterfaceMutationResult]{Results: results, JobID: jobIDFromGen(jobID)}
}
