# Update

Update the systems DNS config:

```bash
$ osapi client network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --search-domains "foo.bar,baz.qux" \
    --interface-name eth0

  ┏━━━━━━━━━━┳━━━━━━━━┳━━━━━━━┓
  ┃ HOSTNAME ┃ STATUS ┃ ERROR ┃
  ┣━━━━━━━━━━╋━━━━━━━━╋━━━━━━━┫
  ┃ server1  ┃ ok     ┃       ┃
  ┗━━━━━━━━━━┻━━━━━━━━┻━━━━━━━┛
```

When targeting all hosts, a confirmation prompt is shown first:

```bash
$ osapi client network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --search-domains "foo.bar" \
    --interface-name eth0 \
    --target _all
This will modify DNS on ALL hosts. Continue? [y/N] y

  ┏━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━┓
  ┃ HOSTNAME ┃ STATUS ┃ ERROR     ┃
  ┣━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━┫
  ┃ server1  ┃ ok     ┃           ┃
  ┃ server2  ┃ failed ┃ disk full ┃
  ┗━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━┛
```

## Flags

| Flag               | Description                                    | Default  |
| ------------------ | ---------------------------------------------- | -------- |
| `--servers`        | List of DNS server IP addresses                | one of\* |
| `--search-domains` | List of DNS search domains                     | one of\* |
| `--interface-name` | Name of the network interface to configure DNS | required |

\*At least one of `--servers` or `--search-domains` must be provided.
