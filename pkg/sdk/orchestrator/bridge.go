package orchestrator

import (
	"encoding/json"
	"fmt"

	client "github.com/retr0h/osapi/pkg/sdk/client"
)

// jsonUnmarshalFn is the JSON unmarshal function (injectable for testing).
var jsonUnmarshalFn = json.Unmarshal

// StructToMap converts a struct to map[string]any using its JSON tags.
// Returns nil if v is nil or cannot be marshaled.
func StructToMap(
	v any,
) map[string]any {
	if v == nil {
		return nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var m map[string]any
	if err := jsonUnmarshalFn(b, &m); err != nil {
		return nil
	}

	return m
}

// CollectionResult builds a Result from a Collection response.
// It iterates all results, applies the toHostResult mapper to build
// per-host details, and auto-populates HostResult.Data via StructToMap
// when the mapper leaves it nil. Changed is true if any host reported
// a change.
//
// When rawJSON is non-nil, it is unmarshaled into Result.Data to
// provide the full response for downstream consumers (e.g., guards
// or Results.Decode). Pass resp.RawJSON() for this, or nil to skip.
func CollectionResult[T any](
	col client.Collection[T],
	rawJSON []byte,
	toHostResult func(T) HostResult,
) (*Result, error) {
	hostResults := make([]HostResult, 0, len(col.Results))
	changed := false

	for _, r := range col.Results {
		hr := toHostResult(r)

		if hr.Data == nil {
			hr.Data = StructToMap(r)
		}

		if hr.Changed {
			changed = true
		}

		hostResults = append(hostResults, hr)
	}

	var data map[string]any
	if len(rawJSON) > 0 {
		if err := jsonUnmarshalFn(rawJSON, &data); err != nil {
			return nil, fmt.Errorf("unmarshal response data: %w", err)
		}
	}

	return &Result{
		JobID:       col.JobID,
		Changed:     changed,
		Data:        data,
		HostResults: hostResults,
	}, nil
}
