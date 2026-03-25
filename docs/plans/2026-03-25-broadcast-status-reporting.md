# Broadcast Status Reporting Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop silently dropping failed/skipped agent responses in cron
broadcast operations so every targeted agent appears in results.

**Architecture:** Remove the status filter in the cron list broadcast
handler. Add `hostname` and `error` fields to the `CronEntry` OpenAPI
schema. Update the SDK type, CLI output, and conversion functions.
Only the cron handler needs fixing — all other broadcast handlers
already pass errors through.

**Tech Stack:** Go 1.25, oapi-codegen, testify/suite, gomock

---

## Chunk 1: API Schema + Handler Fix

### Task 1: Add hostname and error fields to CronEntry schema

**Files:**
- Modify: `internal/controller/api/schedule/gen/api.yaml`

- [ ] **Step 1: Add hostname and error to CronEntry**

In the `CronEntry` schema under `components/schemas`, add:

```yaml
    CronEntry:
      type: object
      description: A cron drop-in entry.
      properties:
        hostname:
          type: string
          description: Hostname of the agent that returned this entry.
        error:
          type: string
          description: Error message if the agent failed or was skipped.
        name:
          ...existing fields...
```

- [ ] **Step 2: Regenerate**

Run: `go generate ./internal/controller/api/schedule/gen/...`
Then: `redocly join --prefix-tags-with-info-prop title -o internal/controller/api/gen/api.yaml internal/controller/api/*/gen/api.yaml`
Then: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 3: Verify build**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```bash
git add internal/controller/api/schedule/gen/ \
        internal/controller/api/gen/ \
        pkg/sdk/client/gen/
git commit -m "feat: add hostname and error fields to CronEntry schema"
```

### Task 2: Fix cron broadcast handler to include failed/skipped

**Files:**
- Modify: `internal/controller/api/schedule/cron_list_get.go`

- [ ] **Step 1: Update getNodeScheduleCronBroadcast**

Replace the filtering logic:

```go
// Before:
allResults := make([]gen.CronEntry, 0)
for _, resp := range responses {
    if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
        continue
    }
    allResults = append(allResults, responseToCronEntries(resp)...)
}

// After:
allResults := make([]gen.CronEntry, 0)
for _, resp := range responses {
    if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
        hostname := resp.Hostname
        errMsg := resp.Error
        if errMsg == "" {
            errMsg = string(resp.Status)
        }
        allResults = append(allResults, gen.CronEntry{
            Hostname: &hostname,
            Error:    &errMsg,
        })
        continue
    }
    allResults = append(allResults, responseToCronEntries(resp)...)
}
```

- [ ] **Step 2: Update responseToCronEntries to set hostname**

The function needs to set `Hostname` on each entry from
`resp.Hostname`:

```go
func responseToCronEntries(
    resp *job.Response,
) []gen.CronEntry {
    var entries []cronProv.Entry
    if resp.Data != nil {
        _ = json.Unmarshal(resp.Data, &entries)
    }

    hostname := resp.Hostname

    results := make([]gen.CronEntry, 0, len(entries))
    for _, e := range entries {
        name := e.Name
        object := e.Object
        source := e.Source

        entry := gen.CronEntry{
            Hostname: &hostname,
            Name:     &name,
            Object:   &object,
            Source:    &source,
        }
        // ...existing field mapping...
        results = append(results, entry)
    }

    return results
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./internal/controller/api/schedule/...`

- [ ] **Step 4: Update tests**

In `internal/controller/api/schedule/cron_list_get_public_test.go`,
update existing broadcast test cases to verify:
- Failed responses appear as entries with `Hostname` and `Error` set
- Skipped responses appear as entries with `Hostname` and `Error` set
- Successful responses have `Hostname` set

Add new test case: "when broadcast has mixed success and failure" that
returns 3 responses (1 success, 1 failed, 1 skipped) and verifies
all 3 appear in results.

- [ ] **Step 5: Run tests**

Run: `go test ./internal/controller/api/schedule/... -v -count=1`

- [ ] **Step 6: Commit**

```bash
git add internal/controller/api/schedule/
git commit -m "fix: include failed/skipped agents in cron broadcast results"
```

---

## Chunk 2: SDK + CLI

### Task 3: Update SDK CronEntryResult type

**Files:**
- Modify: `pkg/sdk/client/schedule_types.go`

- [ ] **Step 1: Add Hostname field to CronEntryResult**

```go
type CronEntryResult struct {
    Hostname string `json:"hostname,omitempty"`
    Name     string `json:"name"`
    Object   string `json:"object,omitempty"`
    Schedule string `json:"schedule,omitempty"`
    Interval string `json:"interval,omitempty"`
    Source   string `json:"source,omitempty"`
    User     string `json:"user,omitempty"`
    Error    string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Update cronEntryCollectionFromGen**

Map the new `Hostname` and `Error` fields in the conversion function:

```go
results = append(results, CronEntryResult{
    Hostname: derefString(r.Hostname),
    Name:     derefString(r.Name),
    Object:   derefString(r.Object),
    // ...
    Error:    derefString(r.Error),
})
```

- [ ] **Step 3: Regenerate SDK client**

Run: `go generate ./pkg/sdk/client/gen/...`

- [ ] **Step 4: Update SDK tests**

In `pkg/sdk/client/schedule_public_test.go`, update `TestCronList`
to include hostname and error fields in mock responses and assert
they're mapped correctly.

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/sdk/client/... -v -count=1`

- [ ] **Step 6: Commit**

```bash
git add pkg/sdk/client/
git commit -m "feat: add Hostname and Error to SDK CronEntryResult"
```

### Task 4: Update CLI cron list to use BuildBroadcastTable

**Files:**
- Modify: `cmd/client_node_schedule_cron_list.go`

- [ ] **Step 1: Replace PrintCompactTable with BuildBroadcastTable**

The cron list CLI currently uses `PrintCompactTable` with raw string
rows. Replace with `BuildBroadcastTable` which handles HOSTNAME,
STATUS, and ERROR columns automatically:

```go
results := make([]cli.ResultRow, 0, len(resp.Data.Results))
for _, r := range resp.Data.Results {
    var errPtr *string
    if r.Error != "" {
        errPtr = &r.Error
    }
    schedule := r.Schedule
    if r.Interval != "" {
        schedule = r.Interval
    }
    results = append(results, cli.ResultRow{
        Hostname: r.Hostname,
        Error:    errPtr,
        Fields:   []string{r.Name, r.Source, schedule, r.Object, r.User},
    })
}
headers, rows := cli.BuildBroadcastTable(
    results,
    []string{"NAME", "SOURCE", "SCHEDULE", "OBJECT", "USER"},
)
cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
```

This gives:
- HOSTNAME column automatically (broadcast table adds it)
- STATUS + ERROR columns when any agent failed/skipped
- Same output as all other broadcast commands

- [ ] **Step 2: Verify build**

Run: `go build ./cmd/...`

- [ ] **Step 3: Commit**

```bash
git add cmd/client_node_schedule_cron_list.go
git commit -m "feat: use BuildBroadcastTable for cron list CLI output"
```

---

## Chunk 3: Docs + Verification

### Task 5: Update docs

**Files:**
- Modify: `docs/docs/sidebar/features/cron-management.md`
- Modify: `docs/docs/sidebar/usage/cli/client/node/schedule/cron.md`

- [ ] **Step 1: Update cron feature docs**

Add a note about broadcast visibility — failed/skipped agents now
appear in list results with STATUS and ERROR columns.

- [ ] **Step 2: Update CLI reference docs**

Update the list output example to show HOSTNAME column and a
failed/skipped agent row.

- [ ] **Step 3: Commit**

```bash
git add docs/
git commit -m "docs: update cron docs for broadcast status visibility"
```

### Task 6: Full verification

- [ ] **Step 1: Build**

Run: `go build ./...`

- [ ] **Step 2: Unit tests**

Run: `just go::unit`

- [ ] **Step 3: Lint**

Run: `just go::vet`

---

## Files Modified Summary

| File | Change |
| --- | --- |
| `internal/controller/api/schedule/gen/api.yaml` | Add hostname, error to CronEntry |
| `internal/controller/api/schedule/cron_list_get.go` | Remove filter, add error entries |
| `internal/controller/api/schedule/cron_list_get_public_test.go` | Add mixed broadcast test |
| `pkg/sdk/client/schedule_types.go` | Add Hostname to CronEntryResult |
| `pkg/sdk/client/schedule_public_test.go` | Update CronList test |
| `cmd/client_node_schedule_cron_list.go` | Use BuildBroadcastTable |
| `docs/` | Update cron feature + CLI docs |
