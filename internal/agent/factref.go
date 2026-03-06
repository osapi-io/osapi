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

	"github.com/retr0h/osapi/internal/job"
)

const factPrefix = "@fact."

// ResolveFacts walks all string values in params and replaces @fact.X
// references with values from the agent's cached facts. Returns a new
// map with resolved values, or an error if any reference cannot be resolved.
func ResolveFacts(
	params map[string]any,
	facts *job.FactsRegistration,
	hostname string,
) (map[string]any, error) {
	if params == nil {
		return nil, nil
	}

	result := make(map[string]any, len(params))
	for k, v := range params {
		resolved, err := resolveValue(v, facts, hostname)
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
	facts *job.FactsRegistration,
	hostname string,
) (any, error) {
	switch val := v.(type) {
	case string:
		return resolveString(val, facts, hostname)
	case map[string]any:
		return ResolveFacts(val, facts, hostname)
	case []any:
		return resolveSlice(val, facts, hostname)
	default:
		return v, nil
	}
}

// resolveSlice resolves @fact.X references in each element of a slice.
func resolveSlice(
	s []any,
	facts *job.FactsRegistration,
	hostname string,
) ([]any, error) {
	result := make([]any, len(s))
	for i, v := range s {
		resolved, err := resolveValue(v, facts, hostname)
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
	facts *job.FactsRegistration,
	hostname string,
) (string, error) {
	if !strings.Contains(s, factPrefix) {
		return s, nil
	}

	result := s
	for strings.Contains(result, factPrefix) {
		start := strings.Index(result, factPrefix)
		// Find the end of the reference (next space, quote, or end of string)
		end := len(result)
		for i := start + len(factPrefix); i < len(result); i++ {
			c := result[i]
			if c == ' ' || c == '"' || c == '\'' || c == ',' || c == ';' {
				end = i
				break
			}
		}

		ref := result[start:end]
		key := result[start+len(factPrefix) : end]

		replacement, err := lookupFact(key, facts, hostname)
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
	facts *job.FactsRegistration,
	hostname string,
) (string, error) {
	if facts == nil {
		return "", fmt.Errorf("facts not available")
	}

	switch key {
	case "interface.primary":
		if facts.PrimaryInterface == "" {
			return "", fmt.Errorf("primary interface not set")
		}
		return facts.PrimaryInterface, nil
	case "hostname":
		if hostname == "" {
			return "", fmt.Errorf("hostname not set")
		}
		return hostname, nil
	case "arch":
		if facts.Architecture == "" {
			return "", fmt.Errorf("architecture not set")
		}
		return facts.Architecture, nil
	case "os":
		return "", fmt.Errorf("os fact not available in FactsRegistration")
	case "kernel":
		if facts.KernelVersion == "" {
			return "", fmt.Errorf("kernel version not set")
		}
		return facts.KernelVersion, nil
	case "fqdn":
		if facts.FQDN == "" {
			return "", fmt.Errorf("fqdn not set")
		}
		return facts.FQDN, nil
	default:
		// Check @fact.custom.X pattern
		if strings.HasPrefix(key, "custom.") {
			customKey := key[len("custom."):]
			if facts.Facts == nil {
				return "", fmt.Errorf("custom fact %q not found", customKey)
			}
			val, ok := facts.Facts[customKey]
			if !ok {
				return "", fmt.Errorf("custom fact %q not found", customKey)
			}
			return fmt.Sprintf("%v", val), nil
		}
		return "", fmt.Errorf("unknown fact key %q", key)
	}
}
