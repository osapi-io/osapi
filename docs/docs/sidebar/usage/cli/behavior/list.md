# List

List every behavior this binary has, with its methods. An internal behavior, one
that exists only to be built on by others, is marked `(internal)`.

```bash
$ osapi behavior list

Behaviors (2):
  NAME         GROUP  METHODS                            DESCRIPTION
  managedfile         get, update, delete                Files osapi writes on a host and keeps a record of (internal)
  sysctl              list, get, create, update, delete  Kernel parameters, written to sysctl.d files and applied
```

```bash
$ osapi behavior list --json
[{"name":"managedfile","doc":"Files osapi writes on a host and keeps a record of","methods":["get","update","delete"],"internal":true},{"name":"sysctl","doc":"Kernel parameters, written to sysctl.d files and applied","methods":["list","get","create","update","delete"]}]
```
