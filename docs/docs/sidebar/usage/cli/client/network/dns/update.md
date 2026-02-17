# Update

Update the systems DNS config:

```bash
$ osapi client network dns update --search-domains "foo,bar,baz" --servers "1.1.1.1,2.2.2.2" --interface-name eth1
10:56AM INF network dns put search_domains=foo,bar,baz servers=1.1.1.1,2.2.2.2 response="" status=ok
```

## Flags

| Flag               | Description                                    | Default  |
| ------------------ | ---------------------------------------------- | -------- |
| `--servers`        | List of DNS server IP addresses                | one of\* |
| `--search-domains` | List of DNS search domains                     | one of\* |
| `--interface-name` | Name of the network interface to configure DNS | required |

\*At least one of `--servers` or `--search-domains` must be provided.
