# Validate

Load a config file and the environment the way `osapi start` does, and run the
same checks, without starting anything. Exits 1 when the config is not valid.

```bash
$ osapi config validate -f ./osapi.yaml

Config:
  File:    ./osapi.yaml
  Status:  valid
```

```bash
$ osapi config validate -f ./osapi.yaml --json
{"file":"./osapi.yaml","valid":false,"error":"..."}
```
