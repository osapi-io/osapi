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

package file

import (
	"encoding/json"
	"io/fs"
	"os"
)

// SetMarshalJSON overrides the marshal function for testing.
func SetMarshalJSON(
	fn func(interface{}) ([]byte, error),
) {
	marshalJSON = fn
}

// ResetMarshalJSON restores the default marshal function.
func ResetMarshalJSON() {
	marshalJSON = json.Marshal
}

// SetOwnershipFuncs overrides how a file's ownership is read and how a name is
// resolved, so a test can describe a system it does not have.
func SetOwnershipFuncs(
	own func(string) (int, int, error),
	uid func(string) (int, error),
	gid func(string) (int, error),
) {
	ownershipOf = own
	lookupUID = uid
	lookupGID = gid
}

// ResetOwnershipFuncs restores the real implementations.
func ResetOwnershipFuncs() {
	ownershipOf = realOwnershipOf
	lookupUID = realLookupUID
	lookupGID = realLookupGID
	statFn = os.Stat
}

// SetStatFn overrides how a path is stat'd, so a test can hand back a FileInfo
// from a platform whose stat structure this package cannot read.
func SetStatFn(
	fn func(string) (fs.FileInfo, error),
) {
	statFn = fn
}

// OwnershipMatches exposes the comparison between a file's ownership and the
// ownership a deploy asked for.
func OwnershipMatches(
	path string,
	wantUID int,
	wantGID int,
) (bool, error) {
	return ownershipMatches(path, wantUID, wantGID)
}

// ErrOwnershipUnreadable is the sentinel for a platform that cannot report a
// file's uid and gid.
var ErrOwnershipUnreadable = errOwnershipUnreadable

// OwnershipOf calls the real implementation, whatever the current override.
func OwnershipOf(
	path string,
) (int, int, error) {
	return realOwnershipOf(path)
}

// LookupUID and LookupGroupID call the real name resolvers.
func LookupUID(name string) (int, error)     { return realLookupUID(name) }
func LookupGroupID(name string) (int, error) { return realLookupGID(name) }

// ResolveID exposes the request-name-to-id conversion.
func ResolveID(
	name string,
	lookup func(string) (int, error),
) (int, error) {
	return resolveID(name, lookup)
}

// UnsetID is the id of an owner or group a request did not name.
const UnsetID = unsetID

var (
	realOwnershipOf = ownershipOf
	realLookupUID   = lookupUID
	realLookupGID   = lookupGID
)
