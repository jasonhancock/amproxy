# amproxy - Authenticated Metrics Proxy

[![Go Reference](https://pkg.go.dev/badge/github.com/jasonhancock/amproxy.svg)](https://pkg.go.dev/github.com/jasonhancock/amproxy)
[![Go Report Card](https://goreportcard.com/badge/jasonhancock/amproxy)](https://goreportcard.com/report/jasonhancock/amproxy)

A proxy for Graphite's Carbon service that authenticates messages before passing them on to Carbon or dropping them on the floor.

## What does this do?

Carbon listens on port 2003 and doesn't offer any sort of authentication. Usually this is manageable by firewalling off the service to only allow connections from hosts you trust. The problem is that I want to build a device that my friends/family run on their networks at home (without static IPs) and report metrics to my Carbon server. I could run some sort of dynamic dns client on each device and dynamically manage my firewall, but I don't really want to deal with that or with doing something like mTLS, or using something like MQTT.

Instead, I run Carbon bound to 127.0.0.1:2003 and run amproxy on port 2005 exposed to the internet. The client devices are each given a public/private key pair that can be used to generate signed messages. These signed messages are sent to amproxy which authenticates the message by validating the signature and whether or not the specified metric is authorized for the given key pair, and if so, forwards the metric on to Carbon.

## Configuration

```none
$ amproxy server --help
Starts the server

Usage:
  amproxy server [flags]

Flags:
      --addr string          The interface and port to bind the server to. Can be set with the ADDR env variable (default "127.0.0.1:2005")
      --auth-file string     The path to the auth file. Can be set with the AUTH_FILE env variable (default "/etc/amproxy/auth_file.yaml")
      --carbon-addr string   The address of the carbon server. Can be set with the CARBON_ADDR env variable (default "127.0.0.1:2003")
  -h, --help                 help for server
      --log-format string    The format of log messages. (logfmt|json) (default "logfmt")
      --log-level string     Log level (all|err|warn|info|debug (default "info")
      --skew duration        The amount of clock skew tolerated. Can be set with the MAX_SKEW env variable (default 5m0s)
```

## Auth File Format

```yaml
---
apikeys:
  my_public_key:
    secret_key: my_secret_key
    metrics:
    - metric1
    - metric2
  my_public_key2:
    secret_key: my_secret_key2
    metrics:
    - metric3
    - metric4
```

In the example above, `my_public_key` is authorized for `metric1` and `metric2` and uses the `my_secret_key` private key.

If the `AUTH_FILE` is updated on disk, it will automatically get reloaded within 60 seconds.

## Protocol

Messages going over the wire are in the form:

```none
metric value timestamp public_key base64_signature
```

### Example

```none
metric = foo
value = 1234
timestamp = 1425059762
public_key = my_public_key
secret_key = my_secret_key
```

The message for which we will generate the signature becomes

```none
foo 1234 1425059762 my_public_key
```

We can generate a signature:

```shell
KEY_PRIVATE=my_secret_key amproxy client signature "foo 1234 1425059762 my_public_key"
```

Which outputs the following:

```none
lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=
```

The message going over the wire becomes:

```none
foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=
```

## Testing

To start a graphite/carbon stack, run:

```shell
docker-compose up
```

Then access the graphite webui at [localhost:8080](http://localhost:8080). Username and password are both `root`.

Start the server in another terminal:

```shell
AUTH_FILE=packaging/redhat/auth_file.yaml go run main.go server
```

Start the test client in another terminal:

```
KEY_PUBLIC=my_public_key KEY_PRIVATE=my_secret_key go run main.go client test-client
```

The test-client will send a metric named `metric1` every 60 seconds with a random value between 30 and 100.
