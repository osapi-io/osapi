# Setting up NATS and the controller

The docs' section Operate has the steps in order: plan, NATS, the controller,
join hosts. This page has the commands of each.

## NATS

osapi's own NATS server runs in operator mode and serves TLS by default: it
makes its credentials and its certificate at its first start, and a controller
or an agent on the same host uses them with no setting.

```sh
osapi nats start                                  # osapi's own NATS server
osapi nats permissions get --role agent --json    # the subjects of a role: controller, agent, setup, join
osapi nats setup --json                           # make osapi's streams and buckets on a server osapi does not run
osapi nats key create --json                      # the key that encrypts the store on the disk
```

For a NATS server that osapi does not run, make the credentials of each role,
then give the administrator the server's part (`--print`):

```sh
osapi nats auth operator create --out /etc/osapi/nats --print --json   # recommended: JWTs that expire and renew
osapi nats auth password create --out /etc/osapi/nats --print --json   # static users
osapi nats auth nkey create --out /etc/osapi/nats --print --json       # nkeys; a join then goes over HTTPS
```

Each writes `osapi-nats.yaml`, the part of the config to merge. Keep
`operator.nk` and `account.nk` offline.

## The controller

```sh
osapi controller start
osapi controller key fingerprint --json                    # the key that each agent gets at its join
osapi controller secret key create --vault default --json  # the key of a vault, before the first start
osapi controller secret key rotate --vault default --json
```

## The config

```sh
osapi config validate -f /etc/osapi/osapi.yaml --json   # check a file as start checks it
osapi config list --json                                # each setting, its value and where it came from
osapi config get --key controller.api.port --json
osapi config default --json                             # the config at its defaults
osapi config spec --json                                # the parts of the config file
```

An environment variable sets each key: `OSAPI_`, then the key in capitals with
`_` for each dot. It has priority over the file.
