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

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/retr0h/osapi/internal/job"
)

// stringPtrOrNil returns a pointer to s if non-empty, otherwise nil.
func stringPtrOrNil(
	s string,
) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ptrToSlice dereferences a *[]string, returning nil if the pointer is nil.
func ptrToSlice(
	p *[]string,
) []string {
	if p == nil {
		return nil
	}
	return *p
}

// envSliceToMap converts a slice of "KEY=VALUE" strings to a map.
func envSliceToMap(
	env *[]string,
) map[string]string {
	if env == nil {
		return nil
	}
	m := make(map[string]string, len(*env))
	for _, e := range *env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

// parsePortMappings converts a slice of "host:container" strings to PortMapping.
func parsePortMappings(
	ports *[]string,
) []job.PortMapping {
	if ports == nil {
		return nil
	}
	var mappings []job.PortMapping
	for _, p := range *ports {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) == 2 {
			host, errH := strconv.Atoi(parts[0])
			container, errC := strconv.Atoi(parts[1])
			if errH == nil && errC == nil {
				mappings = append(mappings, job.PortMapping{
					Host:      host,
					Container: container,
				})
			}
		}
	}
	return mappings
}

// parseVolumeMappings converts a slice of "host:container" strings to VolumeMapping.
func parseVolumeMappings(
	volumes *[]string,
) []job.VolumeMapping {
	if volumes == nil {
		return nil
	}
	var mappings []job.VolumeMapping
	for _, v := range *volumes {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) == 2 {
			mappings = append(mappings, job.VolumeMapping{
				Host:      parts[0],
				Container: parts[1],
			})
		}
	}
	return mappings
}

// portMappingsToStrings converts PortMapping slice to "host:container" strings.
func portMappingsToStrings(
	ports []struct {
		Host      int `json:"host"`
		Container int `json:"container"`
	},
) []string {
	if len(ports) == 0 {
		return nil
	}
	result := make([]string, 0, len(ports))
	for _, p := range ports {
		result = append(result, fmt.Sprintf("%d:%d", p.Host, p.Container))
	}
	return result
}

// volumeMappingsToStrings converts VolumeMapping slice to "host:container" strings.
func volumeMappingsToStrings(
	volumes []struct {
		Host      string `json:"host"`
		Container string `json:"container"`
	},
) []string {
	if len(volumes) == 0 {
		return nil
	}
	result := make([]string, 0, len(volumes))
	for _, v := range volumes {
		result = append(result, fmt.Sprintf("%s:%s", v.Host, v.Container))
	}
	return result
}
