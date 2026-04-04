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

package dns

// ExportGenerateDNSNetplanYAML exposes generateDNSNetplanYAML for testing.
func ExportGenerateDNSNetplanYAML(
	interfaceName string,
	ifaceSection string,
	servers []string,
	searchDomains []string,
) []byte {
	return generateDNSNetplanYAML(interfaceName, ifaceSection, servers, searchDomains)
}

// ExportNetplanSectionForType exposes netplanSectionForType for testing.
func ExportNetplanSectionForType(
	ifaceType string,
) string {
	return netplanSectionForType(ifaceType)
}

// ExportDNSNetplanPath exposes dnsNetplanPath for testing.
func ExportDNSNetplanPath() string {
	return dnsNetplanPath()
}

// ExportResolvePrimaryInterface exposes resolvePrimaryInterface for testing.
func (d *Debian) ExportResolvePrimaryInterface(
	interfaceName string,
) string {
	return d.resolvePrimaryInterface(interfaceName)
}
