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

// Package iface provides network interface configuration management via Netplan.
package iface

import "context"

// Provider manages network interface configuration via Netplan.
type Provider interface {
	// List returns all managed network interface configurations.
	List(ctx context.Context) ([]InterfaceEntry, error)
	// Get returns a single interface configuration by name.
	Get(ctx context.Context, name string) (*InterfaceEntry, error)
	// Create deploys a new interface configuration via Netplan.
	Create(ctx context.Context, entry InterfaceEntry) (*InterfaceResult, error)
	// Update redeploys an existing interface configuration via Netplan.
	Update(ctx context.Context, entry InterfaceEntry) (*InterfaceResult, error)
	// Delete removes an interface configuration via Netplan.
	Delete(ctx context.Context, name string) (*InterfaceResult, error)
}

// InterfaceEntry represents a network interface. For list/get operations,
// the read-only fields (IPv4, IPv6, MAC, Family) are populated from the
// system. For create/update, the config fields (DHCP, Addresses, Gateway,
// etc.) are used to generate Netplan YAML.
type InterfaceEntry struct {
	Name       string   `json:"name"`
	IPv4       string   `json:"ipv4,omitempty"`
	IPv6       string   `json:"ipv6,omitempty"`
	MAC        string   `json:"mac,omitempty"`
	Family     string   `json:"family,omitempty"`
	DHCP4      *bool    `json:"dhcp4,omitempty"`
	DHCP6      *bool    `json:"dhcp6,omitempty"`
	Addresses  []string `json:"addresses,omitempty"`
	Gateway4   string   `json:"gateway4,omitempty"`
	Gateway6   string   `json:"gateway6,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	MACAddress string   `json:"mac_address,omitempty"`
	WakeOnLAN  *bool    `json:"wakeonlan,omitempty"`
	Primary    bool     `json:"primary,omitempty"`
	Managed    bool     `json:"managed,omitempty"`
}

// InterfaceResult is the outcome of a create/update/delete operation.
type InterfaceResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}
