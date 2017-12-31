# ipcalc

[![Build Status](https://travis-ci.org/hazaelsan/ipcalc.svg?branch=master)](https://travis-ci.org/hazaelsan/ipcalc)
[![GoDoc](https://godoc.org/github.com/hazaelsan/ipcalc?status.svg)](https://godoc.org/github.com/hazaelsan/ipcalc)

IP arithmetic utilities.

## Package ipcalc

This package provides basic IP arithmetic utilities, it supports both IPv4/IPv6 addresses.

```go
// Get the next IP address
next := NextIP(net.ParseIP("192.0.2.1")) // 192.0.2.2
```

Note that IP wrapping is entirely possible in most functions
```go
next := NextIP(net.ParseIP("255.255.255.255")) // 0.0.0.0
```

## Package wildcard

This package provides utilities for working with [Wildcard
Masks](https://en.wikipedia.org/wiki/Wildcard_mask), commonly used in Cisco IOS devices.

Wildcard masks are often thought of as "inverse subnet masks", there are some differences:

* Matching logic is reversed from subnet masks, meaning
  * `0` means the equivalent bit must match
  * `1` means the equivalent bit does not matter
* 0s and 1s need not be contiguous

Examples below, format is IP/Wildcard:
  * `192.0.2.0/0.0.0.255` matches `192.0.2.*`
  * `192.0.2.10/0.0.255.0` matches `192.0.*.10`
  * `192.0.2.1/0.0.255.254` matches `192.0.*.{1,3,5,7,...,255}`

Use `ipcalc.Complement` to convert a subnet mask to its wildcard counterpart.
