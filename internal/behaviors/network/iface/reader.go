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
	"net"
	"slices"
	"sort"
	"strconv"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/network"
	"github.com/osapi-io/gohai/pkg/gohai/source"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// reader reads the interfaces through gohai: /sys/class/net and ip on
// Linux, ifconfig and netstat on macOS. It does not write them: a platform
// with no writer reports skipped.
type reader struct {
	src source.Sources
}

func newReader(
	h host.Host,
) (Provider, error) {
	return &reader{src: h.Sources()}, nil
}

// Interfaces is every interface of the host but the loopback, in the
// kernel's order: by index.
func (r *reader) Interfaces(
	ctx context.Context,
) ([]Entry, error) {
	info, err := r.read(ctx)
	if err != nil {
		return nil, err
	}
	ifs := slices.Clone(info.Interfaces)
	sort.SliceStable(ifs, func(a, b int) bool { return ifs[a].Number < ifs[b].Number })
	entries := []Entry{}
	for _, i := range ifs {
		if i.Encapsulation == "Loopback" || i.Name == "lo" || i.Name == "lo0" {
			continue
		}
		entries = append(entries, entry(info, i))
	}

	return entries, nil
}

// Interface is one interface of the host.
func (r *reader) Interface(
	ctx context.Context,
	name string,
) (Entry, error) {
	info, err := r.read(ctx)
	if err != nil {
		return Entry{}, err
	}
	for _, i := range info.Interfaces {
		if i.Name == name {
			return entry(info, i), nil
		}
	}

	return Entry{}, fmt.Errorf("interface %q: %w", name, behavior.ErrNotFound)
}

// Create is unsupported where no writer is declared.
func (r *reader) Create(
	context.Context,
	Config,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Update is unsupported where no writer is declared.
func (r *reader) Update(
	context.Context,
	Config,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// Remove is unsupported where no writer is declared.
func (r *reader) Remove(
	context.Context,
	string,
) (bool, error) {
	return false, behavior.ErrUnsupported
}

// read is gohai's network facts.
func (r *reader) read(
	ctx context.Context,
) (*network.Info, error) {
	out, err := network.NewWith(r.src).Collect(ctx, nil)

	return host.Collected[network.Info](out, err, "read the interfaces")
}

// entry is an interface of info as osapi reports it.
func entry(
	info *network.Info,
	i network.Interface,
) Entry {
	e := Entry{
		Name:      i.Name,
		State:     i.State,
		MAC:       i.MAC,
		MTU:       i.MTU,
		Addresses: []string{},
		Primary:   i.Name == info.DefaultInterface || i.Name == info.DefaultInet6Interface,
	}
	for _, a := range i.Addresses {
		e.Addresses = append(e.Addresses, a.Addr+"/"+strconv.Itoa(a.Prefixlen))
		ip := net.ParseIP(a.Addr)
		if ip == nil || ip.IsLinkLocalUnicast() {
			continue
		}
		if ip.To4() != nil && e.IPv4 == "" {
			e.IPv4 = a.Addr
		}
		if ip.To4() == nil && e.IPv6 == "" {
			e.IPv6 = a.Addr
		}
	}
	for _, rt := range i.Routes {
		if rt.Family == "inet" && rt.Proto == "dhcp" {
			e.DHCP4 = true
		}
	}

	return e
}
