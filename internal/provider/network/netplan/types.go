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

import "context"

// InterfaceProvider manages network interface configuration via Netplan.
type InterfaceProvider interface {
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

// RouteProvider manages route configuration via Netplan.
type RouteProvider interface {
	// List returns all routes from the system routing table.
	List(ctx context.Context) ([]RouteListEntry, error)
	// Get returns the managed routes for a specific interface.
	Get(ctx context.Context, interfaceName string) (*RouteEntry, error)
	// Create deploys new routes for an interface via Netplan.
	Create(ctx context.Context, entry RouteEntry) (*RouteResult, error)
	// Update redeploys routes for an existing interface via Netplan.
	Update(ctx context.Context, entry RouteEntry) (*RouteResult, error)
	// Delete removes managed routes for an interface via Netplan.
	Delete(ctx context.Context, interfaceName string) (*RouteResult, error)
}

// InterfaceEntry represents a managed interface configuration.
type InterfaceEntry struct {
	Name       string   `json:"name"`
	DHCP4      *bool    `json:"dhcp4,omitempty"`
	DHCP6      *bool    `json:"dhcp6,omitempty"`
	Addresses  []string `json:"addresses,omitempty"`
	Gateway4   string   `json:"gateway4,omitempty"`
	Gateway6   string   `json:"gateway6,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	MACAddress string   `json:"mac_address,omitempty"`
	WakeOnLAN  *bool    `json:"wakeonlan,omitempty"`
	Managed    bool     `json:"managed,omitempty"`
}

// InterfaceResult is the outcome of a create/update/delete operation.
type InterfaceResult struct {
	Name    string `json:"name"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// RouteEntry represents managed routes for an interface.
type RouteEntry struct {
	Interface string  `json:"interface"`
	Routes    []Route `json:"routes"`
}

// Route is a single route definition.
type Route struct {
	To     string `json:"to"`
	Via    string `json:"via"`
	Metric int    `json:"metric,omitempty"`
}

// RouteListEntry is a route from the system routing table.
type RouteListEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Mask        string `json:"mask,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Flags       string `json:"flags,omitempty"`
}

// RouteResult is the outcome of a route create/update/delete operation.
type RouteResult struct {
	Interface string `json:"interface"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}
