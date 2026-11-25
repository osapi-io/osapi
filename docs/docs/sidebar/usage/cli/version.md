# Version

Show the version of this binary.

The build of this binary: the version that each component reports, the commit
and the tree it was built from, and the Go that built it.

```bash
osapi version
```

## Examples

```bash
osapi version
osapi version --json
```

## Result

The command prints:

```ansi
$ osapi version
[1;38;2;0;213;255mVersion:[0m
  [38;2;10;143;176mVersion:[0m   v1.4.0
  [38;2;10;143;176mCommit:[0m    e4a8b2dc5f0c7a1b9d3e6f2a8c4b0d7e1f5a9c3b
  [38;2;10;143;176mTree:[0m      clean
  [38;2;10;143;176mBuilt:[0m     2026-01-14T18:22:05Z
  [38;2;10;143;176mBuilt by:[0m  goreleaser
  [38;2;10;143;176mGo:[0m        go1.27.3
  [38;2;10;143;176mPlatform:[0m  linux/amd64
```

With `--json`, the command prints:

```json
{
  "version": "v1.4.0",
  "commit": "e4a8b2dc5f0c7a1b9d3e6f2a8c4b0d7e1f5a9c3b",
  "tree_state": "clean",
  "build_date": "2026-01-14T18:22:05Z",
  "built_by": "goreleaser",
  "go_version": "go1.27.3",
  "platform": "linux/amd64",
  "compiler": "gc",
  "module_checksum": "h1:Qm3a8Zt0n2Jv5k7XqW1yR4sB6cD9eF0gH2iJ3kL5mN8="
}
```

## Global flags

| Flag               | Description                                                                      | Default                 |
| ------------------ | -------------------------------------------------------------------------------- | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                                                     |                         |
| `-j, --json`       | Enable JSON output                                                               |                         |
| `-f, --osapi-file` | Path to config file; without one at the default path, osapi runs on the defaults | `/etc/osapi/osapi.yaml` |

[Output](output.md) tells how the result is printed, and what `--json` gives.
