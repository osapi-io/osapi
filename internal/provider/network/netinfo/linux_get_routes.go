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
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// parseHexIP converts a hex-encoded IP from /proc/net/route to a dotted-quad string.
func parseHexIP(
	hexStr string,
) string {
	if len(hexStr) != 8 {
		return ""
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 4 {
		return ""
	}

	// /proc/net/route stores IPs in little-endian order
	return net.IPv4(b[3], b[2], b[1], b[0]).String()
}

// parseHexMask converts a hex-encoded mask from /proc/net/route to CIDR notation.
func parseHexMask(
	hexStr string,
) string {
	if len(hexStr) != 8 {
		return ""
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 4 {
		return ""
	}

	// Little-endian byte order
	mask := net.IPv4Mask(b[3], b[2], b[1], b[0])
	ones, _ := mask.Size()

	return fmt.Sprintf("/%d", ones)
}

// GetRoutes returns the system routing table by parsing /proc/net/route.
func (l *Linux) GetRoutes() ([]RouteResult, error) {
	rc, err := l.RouteReaderFn()
	if err != nil {
		return nil, fmt.Errorf("failed to read route table: %w", err)
	}
	defer func() { _ = rc.Close() }()

	return parseRoutes(rc)
}

// parseRoutes parses /proc/net/route content into Route structs.
func parseRoutes(
	r io.Reader,
) ([]RouteResult, error) {
	scanner := bufio.NewScanner(r)

	// Skip header line
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty route table")
	}

	var routes []RouteResult

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 8 {
			continue
		}

		metric, _ := strconv.Atoi(fields[6])

		route := RouteResult{
			Interface:   fields[0],
			Destination: parseHexIP(fields[1]),
			Gateway:     parseHexIP(fields[2]),
			Mask:        parseHexMask(fields[7]),
			Metric:      metric,
			Flags:       fields[3],
		}

		routes = append(routes, route)
	}

	return routes, scanner.Err()
}

// GetPrimaryInterface returns the name of the interface used for the default route
// by parsing /proc/net/route.
func (l *Linux) GetPrimaryInterface() (string, error) {
	rc, err := l.RouteReaderFn()
	if err != nil {
		return "", fmt.Errorf("failed to read route table: %w", err)
	}
	defer func() { _ = rc.Close() }()

	scanner := bufio.NewScanner(rc)

	// Skip header
	if !scanner.Scan() {
		return "", fmt.Errorf("empty route table")
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		// Default route has destination 00000000
		if fields[1] == "00000000" {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("no default route found")
}
