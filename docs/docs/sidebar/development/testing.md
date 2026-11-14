---
sidebar_position: 3
---

# Testing

[`CONTRIBUTING.md`][contributing], under "Tests", tells how to run the unit and
integration tests. It also gives the coverage gate and the test conventions of
this repository.

That file is the single source, so this page is a pointer rather than a second
copy of it.

To list all recipes, also the shared recipes:

```bash
just --list
```

The list puts the shared recipes in groups by their prefix, one group for each
tool: `[go]`, `[react]`, `[docs]`, `[md]` and `[just]`. The recipes of this
repository, `test`, `ready` and `generate`, come first.

<!-- prettier-ignore-start -->
[contributing]: https://github.com/osapi-io/osapi/blob/main/CONTRIBUTING.md
<!-- prettier-ignore-end -->
