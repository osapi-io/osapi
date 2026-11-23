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

package sdkmodel

import (
	"strings"
	"unicode"
)

// words are the Go spelling of the words of the spec that Go writes in
// capitals, or as two words.
var words = map[string]string{
	"cpu":            "CPU",
	"dhcp":           "DHCP",
	"dhcp4":          "DHCP4",
	"dhcp6":          "DHCP6",
	"dlq":            "DLQ",
	"dns":            "DNS",
	"fqdn":           "FQDN",
	"gid":            "GID",
	"id":             "ID",
	"ip":             "IP",
	"ipv4":           "IPv4",
	"ipv6":           "IPv6",
	"kv":             "KV",
	"mac":            "MAC",
	"ms":             "MS",
	"mtu":            "MTU",
	"nats":           "NATS",
	"ntp":            "NTP",
	"os":             "OS",
	"pid":            "PID",
	"rss":            "RSS",
	"rtt":            "RTT",
	"sha":            "SHA",
	"sha256":         "SHA256",
	"sshkey":         "SSHKey",
	"uid":            "UID",
	"updatepackages": "UpdatePackages",
	"url":            "URL",
	"utc":            "UTC",
	"uuid":           "UUID",
}

// GoName is the exported Go name of a name of the spec, such as
// SearchDomains for search_domains, or Min15 for 15min: a name cannot start
// with a digit, so the digits go to the end.
func GoName(
	name string,
) string {
	var b strings.Builder
	for _, w := range strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	}) {
		lead := strings.IndexFunc(w, func(r rune) bool { return !unicode.IsDigit(r) })
		if lead > 0 {
			w = w[lead:] + w[:lead]
		}
		if g, ok := words[strings.ToLower(w)]; ok {
			b.WriteString(g)

			continue
		}
		b.WriteString(strings.ToUpper(w[:1]) + w[1:])
	}

	return b.String()
}

// argName is the Go name of an argument, such as interfaceName for
// interface_name, or machineID for machine_id.
func argName(
	name string,
) string {
	g := GoName(name)
	upper := 0
	for upper < len(g) && unicode.IsUpper(rune(g[upper])) {
		upper++
	}
	switch {
	case upper == len(g):
		return strings.ToLower(g)
	case upper > 1:
		// The first word is in capitals, such as IDName: id stays one
		// word.
		upper--
	}

	return strings.ToLower(g[:upper]) + g[upper:]
}
