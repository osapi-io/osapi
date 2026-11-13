---
sidebar_position: 3
---

# Testing

How to run the unit and integration suites, the coverage gate, and the test
conventions this repository follows are in [`CONTRIBUTING.md`][contributing]
under "Testing".

That file is the single source, so this page is a pointer rather than a second
copy of it.

To list every available recipe, including the shared ones:

```bash
just --list
```

The shared recipes are grouped by their prefix, so the list reads by tool:
`[go]`, `[react]`, `[docs]`, `[md]` and `[just]`, after this repository's own
`test`, `ready` and `generate`.

<!-- prettier-ignore-start -->
[contributing]: https://github.com/osapi-io/osapi/blob/main/CONTRIBUTING.md
<!-- prettier-ignore-end -->
