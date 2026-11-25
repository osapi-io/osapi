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

package types

import "slices"

// InterfaceType is the type of a network interface that osapi manages.
type InterfaceType string

// The values of InterfaceType: the types of an interface.
const (
	// InterfaceEthernet is an interface that the host has.
	InterfaceEthernet InterfaceType = "ethernet"
	// InterfaceDummy is a virtual interface that osapi makes.
	InterfaceDummy InterfaceType = "dummy"
)

// Values are the types of an interface.
func (InterfaceType) Values() []InterfaceType {
	return []InterfaceType{InterfaceEthernet, InterfaceDummy}
}

// Valid reports whether i is one of the values.
func (i InterfaceType) Valid() bool {
	return slices.Contains(i.Values(), i)
}

// AddressFamily is the IP address family of an interface.
type AddressFamily string

// The values of AddressFamily: the address families.
const (
	// FamilyInet is IPv4.
	FamilyInet AddressFamily = "inet"
	// FamilyInet6 is IPv6.
	FamilyInet6 AddressFamily = "inet6"
	// FamilyDual is IPv4 and IPv6.
	FamilyDual AddressFamily = "dual"
)

// Values are the address families.
func (AddressFamily) Values() []AddressFamily {
	return []AddressFamily{FamilyInet, FamilyInet6, FamilyDual}
}

// Valid reports whether a is one of the values.
func (a AddressFamily) Valid() bool {
	return slices.Contains(a.Values(), a)
}
