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

package validation

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

// AgentTarget holds the routing-relevant fields of an active worker.
type AgentTarget struct {
	Hostname string
	Labels   map[string]string
}

// WorkerLister returns active workers with their hostnames and labels.
type WorkerLister func(ctx context.Context) ([]AgentTarget, error)

var workerLister WorkerLister

// labelSegmentRe matches NATS subject-safe segments (same as job.labelSegmentRegex).
var labelSegmentRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var (
	cacheMu       sync.Mutex
	cachedWorkers []AgentTarget
	cacheExpiry   time.Time
	cacheTTL      = 5 * time.Second
)

// RegisterTargetValidator registers the valid_target custom validator and
// sets the WorkerLister it uses. Call this at API server startup after the
// job client is created, or in test SetupSuite to inject a mock.
func RegisterTargetValidator(
	lister WorkerLister,
) {
	// Cannot error: tag is non-empty and function is non-nil.
	_ = instance.RegisterValidation("valid_target", validTarget)
	workerLister = lister
	// Reset cache so the new lister is used immediately.
	cacheMu.Lock()
	cacheExpiry = time.Time{}
	cacheMu.Unlock()
}

// getWorkers returns the cached worker list, refreshing from NATS when
// the cache has expired.
func getWorkers() ([]AgentTarget, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if time.Now().Before(cacheExpiry) {
		return cachedWorkers, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	workers, err := workerLister(ctx)
	if err != nil {
		return nil, err
	}

	cachedWorkers = workers
	cacheExpiry = time.Now().Add(cacheTTL)

	return workers, nil
}

// validTarget checks whether the target is a valid routing pattern
// (_any, _all), a label matching an active worker, or a direct hostname.
func validTarget(fl validator.FieldLevel) bool {
	target := fl.Field().String()

	if target == "_any" || target == "_all" {
		return true
	}

	if workerLister == nil {
		return false
	}

	// Label target: validate format then check against active workers.
	if i := strings.IndexByte(target, ':'); i > 0 && i < len(target)-1 {
		key := target[:i]
		value := target[i+1:]
		if !isValidLabelFormat(key, value) {
			return false
		}
		return matchesLabel(key, value)
	}

	// Direct hostname target: check against active workers.
	return matchesHostname(target)
}

// isValidLabelFormat validates that label key and value segments are
// NATS subject-safe.
func isValidLabelFormat(
	key, value string,
) bool {
	if !labelSegmentRe.MatchString(key) {
		return false
	}
	for _, seg := range strings.Split(value, ".") {
		if !labelSegmentRe.MatchString(seg) {
			return false
		}
	}
	return true
}

// matchesLabel checks whether any active worker has a label whose key
// matches and whose value is a prefix match (hierarchical). For example,
// target "group:web" matches a worker with label group=web.dev.us-east.
func matchesLabel(
	key, value string,
) bool {
	workers, err := getWorkers()
	if err != nil {
		return false
	}

	for _, w := range workers {
		if wv, ok := w.Labels[key]; ok {
			if wv == value || strings.HasPrefix(wv, value+".") {
				return true
			}
		}
	}

	return false
}

// matchesHostname checks whether any active worker has the given hostname.
func matchesHostname(
	target string,
) bool {
	workers, err := getWorkers()
	if err != nil {
		return false
	}

	for _, w := range workers {
		if w.Hostname == target {
			return true
		}
	}

	return false
}
