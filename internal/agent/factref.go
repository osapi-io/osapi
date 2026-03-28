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

package agent

import (
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/facts"
	"github.com/retr0h/osapi/internal/job"
)

// ResolveFacts walks all string values in params and replaces @fact.X
// references with values from the agent's cached facts. Returns a new
// map with resolved values, or an error if any reference cannot be resolved.
func ResolveFacts(
	params map[string]any,
	fr *job.FactsRegistration,
	hostname string,
) (map[string]any, error) {
	if params == nil {
		return nil, nil
	}

	result := make(map[string]any, len(params))
	for k, v := range params {
		resolved, err := resolveValue(v, fr, hostname)
		if err != nil {
			return nil, err
		}
		result[k] = resolved
	}

	return result, nil
}

// resolveValue resolves @fact.X references in a single value.
func resolveValue(
	v any,
	fr *job.FactsRegistration,
	hostname string,
) (any, error) {
	switch val := v.(type) {
	case string:
		return resolveString(val, fr, hostname)
	case map[string]any:
		return ResolveFacts(val, fr, hostname)
	case []any:
		return resolveSlice(val, fr, hostname)
	default:
		return v, nil
	}
}

// resolveSlice resolves @fact.X references in each element of a slice.
func resolveSlice(
	s []any,
	fr *job.FactsRegistration,
	hostname string,
) ([]any, error) {
	result := make([]any, len(s))
	for i, v := range s {
		resolved, err := resolveValue(v, fr, hostname)
		if err != nil {
			return nil, err
		}
		result[i] = resolved
	}
	return result, nil
}

// resolveString replaces all @fact.X references in a string.
func resolveString(
	s string,
	fr *job.FactsRegistration,
	hostname string,
) (string, error) {
	if !strings.Contains(s, facts.Prefix) {
		return s, nil
	}

	result := s
	for strings.Contains(result, facts.Prefix) {
		start := strings.Index(result, facts.Prefix)
		// Find the end of the reference (next space, quote, or end of string)
		end := len(result)
		for i := start + len(facts.Prefix); i < len(result); i++ {
			c := result[i]
			if c == ' ' || c == '"' || c == '\'' || c == ',' || c == ';' {
				end = i
				break
			}
		}

		ref := result[start:end]
		key := result[start+len(facts.Prefix) : end]

		replacement, err := lookupFact(key, fr, hostname)
		if err != nil {
			return "", fmt.Errorf("unresolvable fact reference %q: %w", ref, err)
		}

		result = result[:start] + replacement + result[end:]
	}

	return result, nil
}

// lookupFact resolves a fact key to its value.
func lookupFact(
	key string,
	f *job.FactsRegistration,
	hostname string,
) (string, error) {
	if f == nil {
		return "", fmt.Errorf("facts not available")
	}

	switch key {
	case facts.KeyInterfacePrimary:
		if f.PrimaryInterface == "" {
			return "", fmt.Errorf("primary interface not set")
		}
		return f.PrimaryInterface, nil
	case facts.KeyHostname:
		if hostname == "" {
			return "", fmt.Errorf("hostname not set")
		}
		return hostname, nil
	case facts.KeyArch:
		if f.Architecture == "" {
			return "", fmt.Errorf("architecture not set")
		}
		return f.Architecture, nil
	case "os":
		return "", fmt.Errorf("os fact not available in FactsRegistration")
	case facts.KeyKernel:
		if f.KernelVersion == "" {
			return "", fmt.Errorf("kernel version not set")
		}
		return f.KernelVersion, nil
	case facts.KeyFQDN:
		if f.FQDN == "" {
			return "", fmt.Errorf("fqdn not set")
		}
		return f.FQDN, nil
	case facts.KeyContainerized:
		if f.Containerized {
			return "true", nil
		}
		return "false", nil
	default:
		// Check @fact.custom.X pattern
		if facts.IsCustomKey(key) {
			customKey := key[len(facts.CustomPrefix):]
			if f.Facts == nil {
				return "", fmt.Errorf("custom fact %q not found", customKey)
			}
			val, ok := f.Facts[customKey]
			if !ok {
				return "", fmt.Errorf("custom fact %q not found", customKey)
			}
			return fmt.Sprintf("%v", val), nil
		}
		return "", fmt.Errorf("unknown fact key %q", key)
	}
}
