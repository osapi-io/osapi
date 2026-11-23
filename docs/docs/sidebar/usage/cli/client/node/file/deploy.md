# Deploy

Deploy an object of the object store to a path.

Writes the object at the path with its mode and ownership, and keeps a record of
it. A template object, or content_type template, is rendered first with .Facts,
.Vars and .Hostname. A file that holds the content, mode and owner already does
not change and reports changed: false. osapi client file stale list shows the
deployments whose object changed since.

```bash
osapi client node file deploy [flags]
```

## Examples

```bash
# Deploy a file
osapi client node file deploy --target web-01 --object-name example.conf --path /etc/osapi/file-example.conf

# Render a template
osapi client node file deploy --target web-01 --object-name example.conf.tmpl --path /etc/osapi/file-example.conf --vars port=8080

# A file that holds the object
osapi client node file deploy --target web-01 --object-name example.conf --path /etc/osapi/file-example.conf
```

## Result

Deploy a file, on one host:

```ansi
$ osapi client node file deploy --target web-01 --object-name example.conf --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node file deploy --target _all --object-name example.conf --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "hostname": "web-01",
      "path": "/etc/osapi/file-example.conf",
      "sha256": "0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60",
      "status": "ok"
    }
  ]
}
```

Render a template, on one host:

```ansi
$ osapi client node file deploy --target web-01 --object-name example.conf.tmpl --path /etc/osapi/file-example.conf --vars port=8080
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  28ef41e1ff26302a25da25eaa7b2762f2d55d0fdc0631a4f6a8c380d2b8640d4

  1 host: [38;2;103;234;147m1 changed[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node file deploy --target _all --object-name example.conf.tmpl --path /etc/osapi/file-example.conf --vars port=8080
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS   [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  28ef41e1ff26302a25da25eaa7b2762f2d55d0fdc0631a4f6a8c380d2b8640d4
  [1mweb-02[0m    [38;2;103;234;147mchanged[0m  /etc/osapi/file-example.conf  28ef41e1ff26302a25da25eaa7b2762f2d55d0fdc0631a4f6a8c380d2b8640d4

  3 hosts: [38;2;103;234;147m2 changed[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": true,
      "hostname": "web-01",
      "path": "/etc/osapi/file-example.conf",
      "sha256": "28ef41e1ff26302a25da25eaa7b2762f2d55d0fdc0631a4f6a8c380d2b8640d4",
      "status": "ok"
    }
  ]
}
```

A file that holds the object, on one host:

```ansi
$ osapi client node file deploy --target web-01 --object-name example.conf --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60

  1 host: [38;2;103;234;147m1 ok[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m
```

On each host, where one cannot run it:

```ansi
$ osapi client node file deploy --target _all --object-name example.conf --path /etc/osapi/file-example.conf
[1;38;2;0;213;255mResults:[0m
  [38;2;10;143;176mHOSTNAME  [0m[38;2;10;143;176mSTATUS  [0m[38;2;10;143;176mPATH                          [0m[38;2;10;143;176mSHA256[0m
  [1mweb-01[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60
  [1mweb-02[0m    [38;2;103;234;147mok[0m      /etc/osapi/file-example.conf  0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60

  3 hosts: [38;2;103;234;147m2 ok[0m, [38;5;222m1 skipped[0m   [38;2;138;138;138mjob 550e8400-e29b-41d4-a716-446655440000[0m   [38;2;138;138;138mtrace 26e448c9dbe1c4a8d216a98c675079c9[0m

[1;38;2;255;0;255mDetails:[0m
  [1mmac-01[0m  [38;5;222mskipped[0m  operation not supported on this OS family
```

With `--json`, the API's response:

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "changed": false,
      "hostname": "web-01",
      "path": "/etc/osapi/file-example.conf",
      "sha256": "0083dacc561dfd01081d8554269e201f4b60fe597ab9892977bfde81da46ef60",
      "status": "ok"
    }
  ]
}
```

With the default settings, each example does this on the host:

| Example           | On the host                                                       |
| ----------------- | ----------------------------------------------------------------- |
| Deploy a file     | writes `/etc/osapi/file-example.conf`: `listen 8080`              |
| Render a template | writes `/etc/osapi/file-example.conf`: `host test-host port 8080` |

The command fails in these cases:

| Call                                      | Flags                                                                                                                                                                                                                                                                                                                | Fails with   |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| A relative path                           | `--object-name example.conf --path etc/x`                                                                                                                                                                                                                                                                            | startswith   |
| No object                                 | `--path /etc/osapi/file-example.conf`                                                                                                                                                                                                                                                                                | required     |
| A mode that is not octal                  | `--object-name example.conf --path /etc/osapi/file-example.conf --mode rw-r--r--`                                                                                                                                                                                                                                    | file_mode    |
| An object name longer than 255 characters | `--object-name aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa --path /etc/osapi/file-example.conf` | max          |
| A group that is an option                 | `--object-name example.conf --path /etc/osapi/file-example.conf --group -r`                                                                                                                                                                                                                                          | account_name |
| An owner that is an option                | `--object-name example.conf --path /etc/osapi/file-example.conf --owner -r`                                                                                                                                                                                                                                          | account_name |
| A content type that is neither            | `--object-name example.conf --path /etc/osapi/file-example.conf --content-type binary`                                                                                                                                                                                                                               | oneof        |

## Permission

It needs the `file:execute` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag             | Description                                                                                                                                 | Default |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `--content-type` | How the object is written: raw as it is, or template rendered with .Facts, .Vars and .Hostname; the object's own when empty (raw, template) |         |
| `--group`        | Group that owns the file; left as written when empty                                                                                        |         |
| `--mode`         | Permissions in octal, such as 0644; when empty, a file keeps its mode and a new file gets 0644                                              |         |
| `--object-name`  | The object of the object store, from osapi client file list (required)                                                                      |         |
| `--owner`        | User that owns the file; left as written when empty                                                                                         |         |
| `--path`         | Absolute path of the file on the host (required)                                                                                            |         |
| `-T, --target`   | Target: _all, _any, hostname, or label (group:web.dev)                                                                                      | `_all`  |
| `--vars`         | Variables of a template, as .Vars, each as key=value                                                                                        |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
