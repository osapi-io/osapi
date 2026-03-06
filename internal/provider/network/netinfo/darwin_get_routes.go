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

package netinfo

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// GetRoutes returns the macOS routing table by parsing `netstat -rn` output.
//
// Expected format:
//
//	Routing tables
//	Internet:
//	Destination        Gateway            Flags     Netif Expire
//	default            192.168.1.1        UGScg       en0
//	127                127.0.0.1          UCS         lo0
//	192.168.1/24       link#6             UCS         en0
func (d *Darwin) GetRoutes() ([]RouteResult, error) {
	rc, err := d.RouteReaderFn()
	if err != nil {
		return nil, fmt.Errorf("failed to read route table: %w", err)
	}
	defer func() { _ = rc.Close() }()

	return parseDarwinRoutes(rc)
}

// parseDarwinRoutes parses BSD-style `netstat -rn` output.
func parseDarwinRoutes(
	r io.Reader,
) ([]RouteResult, error) {
	scanner := bufio.NewScanner(r)

	// Find the "Internet:" section and skip to the column header row
	inIPv4Section := false
	foundHeader := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "Internet:" {
			inIPv4Section = true
			continue
		}

		if inIPv4Section && strings.HasPrefix(line, "Destination") {
			foundHeader = true
			break
		}

		// Stop if we hit Internet6: before finding the header
		if inIPv4Section && line == "Internet6:" {
			break
		}
	}

	if !foundHeader {
		return nil, fmt.Errorf("no IPv4 routing table found in netstat output")
	}

	var routes []RouteResult

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Stop at the next section (e.g., Internet6:)
		if line == "" || line == "Internet6:" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		route := RouteResult{
			Destination: fields[0],
			Gateway:     fields[1],
			Flags:       fields[2],
			Interface:   fields[3],
		}

		routes = append(routes, route)
	}

	return routes, scanner.Err()
}

// GetPrimaryInterface returns the name of the interface used for the default
// route from macOS `netstat -rn` output.
func (d *Darwin) GetPrimaryInterface() (string, error) {
	routes, err := d.GetRoutes()
	if err != nil {
		return "", err
	}

	for _, route := range routes {
		if route.Destination == "default" {
			return route.Interface, nil
		}
	}

	return "", fmt.Errorf("no default route found")
}
