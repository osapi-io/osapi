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

package container

import "github.com/retr0h/osapi/internal/job"

// ExportStringPtrOrNil exposes the private stringPtrOrNil for testing.
func ExportStringPtrOrNil(
	s string,
) *string {
	return stringPtrOrNil(s)
}

// ExportPtrToSlice exposes the private ptrToSlice for testing.
func ExportPtrToSlice(
	s *[]string,
) []string {
	return ptrToSlice(s)
}

// ExportEnvSliceToMap exposes the private envSliceToMap for testing.
func ExportEnvSliceToMap(
	s *[]string,
) map[string]string {
	return envSliceToMap(s)
}

// ExportParsePortMappings exposes the private parsePortMappings for testing.
func ExportParsePortMappings(
	s *[]string,
) []job.PortMapping {
	return parsePortMappings(s)
}

// ExportParseVolumeMappings exposes the private parseVolumeMappings for testing.
func ExportParseVolumeMappings(
	s *[]string,
) []job.VolumeMapping {
	return parseVolumeMappings(s)
}

// ExportPortMappingsToStrings exposes the private portMappingsToStrings for testing.
func ExportPortMappingsToStrings(
	ports []struct {
		Host      int `json:"host"`
		Container int `json:"container"`
	},
) []string {
	return portMappingsToStrings(ports)
}

// ExportVolumeMappingsToStrings exposes the private volumeMappingsToStrings for testing.
func ExportVolumeMappingsToStrings(
	volumes []struct {
		Host      string `json:"host"`
		Container string `json:"container"`
	},
) []string {
	return volumeMappingsToStrings(volumes)
}

// ExportNilIfEmptyStrSlice exposes the private nilIfEmptyStrSlice for testing.
func ExportNilIfEmptyStrSlice(
	s []string,
) *[]string {
	return nilIfEmptyStrSlice(s)
}

// ExportInt64PtrOrNil exposes the private int64PtrOrNil for testing.
func ExportInt64PtrOrNil(
	v int64,
) *int64 {
	return int64PtrOrNil(v)
}
