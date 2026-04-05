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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/provider/network/netplan"
	"github.com/retr0h/osapi/internal/provider/network/netplan/iface"
)

// filteredRouteTypes are route types that represent kernel-internal
// entries and should be excluded from the user-visible route list.
var filteredRouteTypes = map[string]bool{
	"local":     true,
	"broadcast": true,
	"anycast":   true,
	"multicast": true,
}

// filteredRouteScopes are route scopes for kernel-internal entries.
var filteredRouteScopes = map[string]bool{
	"host": true,
}

const (
	netplanDir      = "/etc/netplan"
	interfacePrefix = "osapi-"
)

// routeFilePath returns the Netplan route config file path for an interface.
func routeFilePath(
	interfaceName string,
) string {
	return netplanDir + "/" + interfacePrefix + interfaceName + "-routes.yaml"
}

// List returns all routes from the system routing table, excluding
// kernel-internal entries (local, broadcast, anycast, multicast types
// and host-scoped routes).
func (d *Debian) List(
	_ context.Context,
) ([]ListEntry, error) {
	status, err := netplan.GetStatus(d.execManager)
	if err != nil {
		return nil, fmt.Errorf("netplan route list: %w", err)
	}

	var result []ListEntry

	for ifaceName, iface := range status {
		for _, r := range iface.Routes {
			if filteredRouteTypes[r.Type] || filteredRouteScopes[r.Scope] {
				continue
			}

			result = append(result, ListEntry{
				Destination: r.To,
				Gateway:     r.Via,
				Interface:   ifaceName,
				Metric:      r.Metric,
			})
		}
	}

	return result, nil
}

// Get returns the managed routes for a specific interface by reading
// the route metadata from the file-state KV store.
func (d *Debian) Get(
	ctx context.Context,
	interfaceName string,
) (*Entry, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("netplan route get: interface name must not be empty")
	}

	path := routeFilePath(interfaceName)

	stateKey := file.BuildStateKey(d.hostname, path)

	kvEntry, err := d.stateKV.Get(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf("netplan route %q: not found", interfaceName)
	}

	var state struct {
		UndeployedAt string            `json:"undeployed_at,omitempty"`
		Metadata     map[string]string `json:"metadata,omitempty"`
	}

	if unmarshalErr := json.Unmarshal(kvEntry.Value(), &state); unmarshalErr != nil {
		return nil, fmt.Errorf("netplan route get: unmarshal state: %w", unmarshalErr)
	}

	if state.UndeployedAt != "" {
		return nil, fmt.Errorf("netplan route %q: not found", interfaceName)
	}

	routesJSON, ok := state.Metadata["routes"]
	if !ok {
		return nil, fmt.Errorf("netplan route get: no route metadata for %q", interfaceName)
	}

	var routes []Route
	if unmarshalErr := json.Unmarshal([]byte(routesJSON), &routes); unmarshalErr != nil {
		return nil, fmt.Errorf("netplan route get: unmarshal routes: %w", unmarshalErr)
	}

	return &Entry{
		Interface: interfaceName,
		Routes:    routes,
	}, nil
}

// Create deploys new routes for an interface via Netplan. Fails if a
// managed route file already exists for the interface, or if any route
// targets the default gateway.
func (d *Debian) Create(
	ctx context.Context,
	entry Entry,
) (*Result, error) {
	if err := iface.ValidateInterfaceName(entry.Interface); err != nil {
		return nil, fmt.Errorf("netplan route create: %w", err)
	}

	if containsDefaultRoute(entry.Routes) {
		return nil, fmt.Errorf(
			"netplan route create: default route (0.0.0.0/0, ::/0, default) not allowed",
		)
	}

	path := routeFilePath(entry.Interface)

	// Fail if the managed file already exists on disk.
	if _, statErr := d.fs.Stat(path); statErr == nil {
		return nil, fmt.Errorf(
			"netplan route create: %q already managed",
			entry.Interface,
		)
	}

	ifaceSection := netplan.SectionForInterface(d.execManager, entry.Interface)
	content := generateRouteYAML(entry, ifaceSection)

	metadata, err := buildRouteMetadata(entry)
	if err != nil {
		return nil, fmt.Errorf("netplan route create: %w", err)
	}

	changed, applyErr := netplan.ApplyConfig(
		ctx,
		d.logger,
		d.fs,
		d.stateKV,
		d.execManager,
		d.hostname,
		path,
		content,
		metadata,
	)
	if applyErr != nil {
		return nil, fmt.Errorf("netplan route create: %w", applyErr)
	}

	d.logger.Info(
		"netplan route created",
		slog.String("interface", entry.Interface),
		slog.Bool("changed", changed),
	)

	return &Result{
		Interface: entry.Interface,
		Changed:   changed,
	}, nil
}

// Update redeploys routes for an existing interface via Netplan. Fails
// if no managed route file exists for the interface, or if any route
// targets the default gateway.
func (d *Debian) Update(
	ctx context.Context,
	entry Entry,
) (*Result, error) {
	if err := iface.ValidateInterfaceName(entry.Interface); err != nil {
		return nil, fmt.Errorf("netplan route update: %w", err)
	}

	if containsDefaultRoute(entry.Routes) {
		return nil, fmt.Errorf(
			"netplan route update: default route (0.0.0.0/0, ::/0, default) not allowed",
		)
	}

	path := routeFilePath(entry.Interface)

	// Fail if the managed file does not exist on disk.
	if _, statErr := d.fs.Stat(path); statErr != nil {
		return nil, fmt.Errorf(
			"netplan route update: %q not managed",
			entry.Interface,
		)
	}

	ifaceSection := netplan.SectionForInterface(d.execManager, entry.Interface)
	content := generateRouteYAML(entry, ifaceSection)

	metadata, err := buildRouteMetadata(entry)
	if err != nil {
		return nil, fmt.Errorf("netplan route update: %w", err)
	}

	changed, applyErr := netplan.ApplyConfig(
		ctx,
		d.logger,
		d.fs,
		d.stateKV,
		d.execManager,
		d.hostname,
		path,
		content,
		metadata,
	)
	if applyErr != nil {
		return nil, fmt.Errorf("netplan route update: %w", applyErr)
	}

	d.logger.Info(
		"netplan route updated",
		slog.String("interface", entry.Interface),
		slog.Bool("changed", changed),
	)

	return &Result{
		Interface: entry.Interface,
		Changed:   changed,
	}, nil
}

// Delete removes managed routes for an interface via Netplan.
// Returns an error if the routes are not managed by OSAPI.
func (d *Debian) Delete(
	ctx context.Context,
	interfaceName string,
) (*Result, error) {
	if interfaceName == "" {
		return nil, fmt.Errorf("netplan route delete: interface name must not be empty")
	}

	path := routeFilePath(interfaceName)

	if _, err := d.fs.Stat(path); err != nil {
		return nil, fmt.Errorf(
			"netplan route delete: %q not managed",
			interfaceName,
		)
	}

	changed, err := netplan.RemoveConfig(
		ctx,
		d.logger,
		d.fs,
		d.stateKV,
		d.execManager,
		d.hostname,
		path,
	)
	if err != nil {
		return nil, fmt.Errorf("netplan route delete: %w", err)
	}

	if changed {
		d.logger.Info(
			"netplan route deleted",
			slog.String("interface", interfaceName),
		)
	}

	return &Result{
		Interface: interfaceName,
		Changed:   changed,
	}, nil
}

// containsDefaultRoute returns true if any route targets the default
// gateway (0.0.0.0/0, ::/0, or "default").
func containsDefaultRoute(
	routes []Route,
) bool {
	for _, r := range routes {
		if r.To == "0.0.0.0/0" || r.To == "::/0" || r.To == "default" {
			return true
		}
	}

	return false
}

// generateRouteYAML builds a Netplan YAML configuration for the given
// route entry.
func generateRouteYAML(
	entry Entry,
	ifaceSection string,
) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, "network:\n")
	fmt.Fprintf(&b, "  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n", ifaceSection)
	fmt.Fprintf(&b, "    %s:\n", entry.Interface)
	fmt.Fprintf(&b, "      routes:\n")

	for _, r := range entry.Routes {
		fmt.Fprintf(&b, "        - to: %s\n", r.To)
		fmt.Fprintf(&b, "          via: %s\n", r.Via)

		if r.Metric > 0 {
			fmt.Fprintf(&b, "          metric: %d\n", r.Metric)
		}
	}

	return []byte(b.String())
}

// buildRouteMetadata serializes route data into the metadata map for
// storage in the file-state KV.
func buildRouteMetadata(
	entry Entry,
) (map[string]string, error) {
	routesJSON, err := marshalJSON(entry.Routes)
	if err != nil {
		return nil, fmt.Errorf("marshal routes: %w", err)
	}

	return map[string]string{
		"interface": entry.Interface,
		"routes":    string(routesJSON),
	}, nil
}

