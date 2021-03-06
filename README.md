# amproxy - Authenticated Metrics Proxy

[![GoDoc](https://godoc.org/github.com/jasonhancock/amproxy?status.svg)](https://godoc.org/github.com/jasonhancock/amproxy)
[![Build Status](https://travis-ci.org/jasonhancock/amproxy.svg?branch=master)](https://travis-ci.org/jasonhancock/amproxy)
[![Go Report Card](https://goreportcard.com/badge/jasonhancock/amproxy)](https://goreportcard.com/report/jasonhancock/amproxy)

A proxy for Graphite's Carbon service that authenticates messages before passing
them on to Carbon or dropping them on the floor.

This is a prototype and is just an example of what could be possible. It is
quite literally the first code beyond a "Hello World" that I've written in Go.

## What does this do?

Carbon listens on port 2003 and doesn't offer any sort of authentication.
Usually this is manageable by firewalling off the service to only allow
connections from hosts you trust. The problem is that I want to build a device
that my friends/family run on their networks at home (without static IPs) and
report metrics to my Carbon server. I could run some sort of dynamic dns client
on each device and dynamically manage my firewall, but I don't really want to
deal with that.

Instead, I run Carbon bound to 127.0.0.1:2003 and run amproxy on port 2005
exposed to the internet. The client devices are each given a public/private key
pair that can be used to generate signed messages. These signed messages are
sent to amproxy which authenticates the message by validating the signature
and whether or not the specified metric is authorized for the given key pair,
and if so, forwards the metric on to Carbon.

## Configuration

All configuration is done via flags (will be updated for env vars soon):

$ ./amproxy --help
Usage of ./amproxy:
  -addr string
        interface/port to bind to (default ":2005")
  -auth-file string
        Location of auth file (default "/etc/amproxy/auth_file.yaml")
  -carbon-addr string
        Carbon address:port (default "127.0.0.1:2003")
  -skew float
        amount of clock skew tolerated in seconds (default 300)

## Auth File Format

```
---
apikeys:
  my_public_key:
    secret_key: my_secret_key
    metrics:
      metric1: 1
      metric2: 1
  my_public_key2:
    secret_key: my_secret_key2
    metrics:
      metric3: 1
      metric4: 1
```

In the example above, my_public_key is authorized for metric1 and metric2 and
uses the `my_secret_key` private key.

If the AUTH_FILE is updated on disk, it will automatically get reloaded within
60 seconds.

## Protocol

Messages going over the wire are in the form:

```
metric value timestamp public_key base64_signature
```

### Example:

```
metric = foo
value = 1234
timestamp = 1425059762
public_key = my_public_key
secret_key = my_secret_key
```

The message for which we will generate the signature becomes

```
foo 1234 1425059762 my_public_key
```

We can generate a signature with some perl code:

```
#!/usr/bin/perl

use strict;
use warnings;
use Digest::SHA qw(hmac_sha256_base64);

my $digest = hmac_sha256_base64('foo 1234 1425059762 my_public_key', 'my_secret_key');

# Fix padding of Base64 digests
while (length($digest) % 4) {
    $digest .= '=';
}

print $digest;
```

Which outputs the following:
```
lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=
```

The message going over the wire becomes:

```
foo 1234 1425059762 my_public_key lT9zOeBVNfTdogqKE5J7p3XWprfu/gOI5D7aWRzjJtc=
```

## Building/Testing

To build in the vagrant environment, do the following:

```
cd /vagrant/src/amproxy/amproxy
go install
```

This will generate the `/vagrant/bin/amproxy` binary. You can then run the binary:

```
AUTH=public_key1:private_key1 /vagrant/bin/amproxy
```

And ship your signed metrics to localhost:2005

## Ideas

This was just a proof of concept. Ideas for the future would be some sort of
pluggable backend to fetch the public/private keypairs from. As I'm still
prototyping, I didn't want to build out a complicated system that tied into
MySQL, Redis, Memcached, or some other backend API.
