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

package iface

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"

	"github.com/retr0h/osapi/internal/provider/network/netplan"
)

const (
	netplanDir      = "/etc/netplan"
	interfacePrefix = "osapi-"
)

// ValidName matches alphanumeric characters and dashes.
var ValidName = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

// interfaceFilePath returns the Netplan config file path for an interface.
func interfaceFilePath(
	name string,
) string {
	return netplanDir + "/" + interfacePrefix + name + ".yaml"
}

// List returns all network interfaces from netplan status.
func (d *Debian) List(
	_ context.Context,
) ([]InterfaceEntry, error) {
	status, err := netplan.GetStatus(d.execManager)
	if err != nil {
		return nil, fmt.Errorf("interface list: %w", err)
	}

	var result []InterfaceEntry

	for name, iface := range status {
		if name == "lo" {
			continue
		}
		dhcp := iface.IsDHCP()

		entry := InterfaceEntry{
			Name:    name,
			IPv4:    iface.IPv4(),
			IPv6:    iface.IPv6(),
			MAC:     iface.MACAddress,
			Family:  iface.AddressFamily(),
			Primary: iface.HasDefaultRoute(),
			DHCP4:   &dhcp,
		}

		result = append(result, entry)
	}

	// Sort by interface index for stable ordering. Interfaces
	// without an index fall back to alphabetical name order.
	sort.Slice(result, func(i, j int) bool {
		iIdx := status[result[i].Name].Index
		jIdx := status[result[j].Name].Index

		if iIdx != jIdx {
			return iIdx < jIdx
		}

		return result[i].Name < result[j].Name
	})

	return result, nil
}

// Get returns a single interface by name.
func (d *Debian) Get(
	_ context.Context,
	name string,
) (*InterfaceEntry, error) {
	if name == "" {
		return nil, fmt.Errorf("interface get: name must not be empty")
	}

	status, err := netplan.GetStatus(d.execManager)
	if err != nil {
		return nil, fmt.Errorf("interface get: %w", err)
	}

	iface, ok := status[name]
	if !ok {
		return nil, fmt.Errorf("interface %q: not found", name)
	}

	dhcp := iface.IsDHCP()

	entry := &InterfaceEntry{
		Name:    name,
		IPv4:    iface.IPv4(),
		IPv6:    iface.IPv6(),
		MAC:     iface.MACAddress,
		Family:  iface.AddressFamily(),
		Primary: iface.HasDefaultRoute(),
		DHCP4:   &dhcp,
	}

	return entry, nil
}

// Create deploys a new Netplan interface configuration file. Fails if
// a managed file already exists for the interface name.
func (d *Debian) Create(
	ctx context.Context,
	entry InterfaceEntry,
) (*InterfaceResult, error) {
	if err := ValidateInterfaceName(entry.Name); err != nil {
		return nil, fmt.Errorf("interface create: %w", err)
	}

	path := interfaceFilePath(entry.Name)

	// Already managed — nothing to do.
	if _, statErr := d.fs.Stat(path); statErr == nil {
		return &InterfaceResult{
			Name:    entry.Name,
			Changed: false,
		}, nil
	}

	ifaceSection := netplan.SectionForInterface(d.execManager, entry.Name)
	content := generateInterfaceYAML(entry, ifaceSection)
	metadata := map[string]string{
		"interface": entry.Name,
	}

	changed, err := netplan.ApplyConfig(
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
	if err != nil {
		return nil, fmt.Errorf("interface create: %w", err)
	}

	d.logger.Info(
		"interface created",
		slog.String("name", entry.Name),
		slog.Bool("changed", changed),
	)

	return &InterfaceResult{
		Name:    entry.Name,
		Changed: changed,
	}, nil
}

// Update redeploys an existing Netplan interface configuration file.
// Fails if no managed file exists for the interface name.
func (d *Debian) Update(
	ctx context.Context,
	entry InterfaceEntry,
) (*InterfaceResult, error) {
	if err := ValidateInterfaceName(entry.Name); err != nil {
		return nil, fmt.Errorf("interface update: %w", err)
	}

	path := interfaceFilePath(entry.Name)

	// Fail if the managed file does not exist on disk.
	if _, statErr := d.fs.Stat(path); statErr != nil {
		return nil, fmt.Errorf(
			"interface update: %q not managed",
			entry.Name,
		)
	}

	ifaceSection := netplan.SectionForInterface(d.execManager, entry.Name)
	content := generateInterfaceYAML(entry, ifaceSection)
	metadata := map[string]string{
		"interface": entry.Name,
	}

	changed, err := netplan.ApplyConfig(
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
	if err != nil {
		return nil, fmt.Errorf("interface update: %w", err)
	}

	d.logger.Info(
		"interface updated",
		slog.String("name", entry.Name),
		slog.Bool("changed", changed),
	)

	return &InterfaceResult{
		Name:    entry.Name,
		Changed: changed,
	}, nil
}

// Delete removes a managed Netplan interface configuration file.
// If no managed file exists, returns Changed: false (idempotent).
func (d *Debian) Delete(
	ctx context.Context,
	name string,
) (*InterfaceResult, error) {
	if name == "" {
		return nil, fmt.Errorf("interface delete: name must not be empty")
	}

	path := interfaceFilePath(name)

	if _, err := d.fs.Stat(path); err != nil {
		return &InterfaceResult{
			Name:    name,
			Changed: false,
		}, nil
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
		return nil, fmt.Errorf("interface delete: %w", err)
	}

	if changed {
		d.logger.Info(
			"interface deleted",
			slog.String("name", name),
		)
	}

	return &InterfaceResult{
		Name:    name,
		Changed: changed,
	}, nil
}

// ValidateInterfaceName checks that the interface name is non-empty and
// matches the allowed pattern (alphanumeric and dashes).
func ValidateInterfaceName(
	name string,
) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}

	if !ValidName.MatchString(name) {
		return fmt.Errorf("name %q contains invalid characters", name)
	}

	return nil
}

// generateInterfaceYAML builds a Netplan YAML configuration for the
// given interface entry. Only non-zero fields are included. The
// ifaceSection parameter specifies the YAML section (ethernets, wifis, etc.).
func generateInterfaceYAML(
	entry InterfaceEntry,
	ifaceSection string,
) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, "network:\n")
	fmt.Fprintf(&b, "  version: 2\n")
	fmt.Fprintf(&b, "  %s:\n", ifaceSection)
	fmt.Fprintf(&b, "    %s:\n", entry.Name)

	if entry.DHCP4 != nil {
		fmt.Fprintf(&b, "      dhcp4: %t\n", *entry.DHCP4)
	}

	if entry.DHCP6 != nil {
		fmt.Fprintf(&b, "      dhcp6: %t\n", *entry.DHCP6)
	}

	if len(entry.Addresses) > 0 {
		fmt.Fprintf(&b, "      addresses:\n")

		for _, addr := range entry.Addresses {
			fmt.Fprintf(&b, "        - %s\n", addr)
		}
	}

	if entry.Gateway4 != "" {
		fmt.Fprintf(&b, "      gateway4: %s\n", entry.Gateway4)
	}

	if entry.Gateway6 != "" {
		fmt.Fprintf(&b, "      gateway6: %s\n", entry.Gateway6)
	}

	if entry.MTU > 0 {
		fmt.Fprintf(&b, "      mtu: %d\n", entry.MTU)
	}

	if entry.MACAddress != "" {
		fmt.Fprintf(&b, "      macaddress: %s\n", entry.MACAddress)
	}

	if entry.WakeOnLAN != nil {
		fmt.Fprintf(&b, "      wakeonlan: %t\n", *entry.WakeOnLAN)
	}

	return []byte(b.String())
}
