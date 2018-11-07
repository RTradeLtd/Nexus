# 🦑 ipfs-orchestrator

The IPFS private network node orchestration and registry service for
[Temporal](https://github.com/RTradeLtd/Temporal), an easy-to-use interface into
distributed and decentralized storage technologies.

[![GoDoc](https://godoc.org/github.com/RTradeLtd/ipfs-orchestrator?status.svg)](https://godoc.org/github.com/RTradeLtd/ipfs-orchestrator)
[![Build Status](https://travis-ci.com/RTradeLtd/ipfs-orchestrator.svg?branch=master)](https://travis-ci.com/RTradeLtd/ipfs-orchestrator)
[![codecov](https://codecov.io/gh/RTradeLtd/ipfs-orchestrator/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/ipfs-orchestrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/ipfs-orchestrator)](https://goreportcard.com/report/github.com/RTradeLtd/ipfs-orchestrator)

## Installation and Usage

```bash
$> go get -u github.com/RTradeLtd/ipfs-orchestrator
```

Releases are also be available from the
[Releases](https://github.com/RTradeLtd/ipfs-orchestrator/releases) page.

```bash
$> ipfs-orchestrator init
$> ipfs-orchestrator daemon --address $MY_HOST
```

Further documentation is available via `ipfs-orchestrator --help`.

## Development

This project requires [Docker CE](https://docs.docker.com/install/#supported-platforms)
and [Go 1.11](https://golang.org/dl/) or above.

To fetch the codebase, clone the repository or use `go get`:

```bash
$> go get github.com/RTradeLtd/ipfs-orchestrator
```

Dependencies can be installed using the provided Makefile:

```bash
$> make   # installs dependencies and builds a binary
```

### Testing

To execute the tests, make sure the Docker daemon is running and run:

```bash
$> make test
```

You can remove leftover assets using `make clean`.

### ctl

An experimental, lightweight controller for the gRPC API is available via the
`ipfs-orchestrator ctl` command, which exposes a client via the
[ctl](https://github.com/bobheadxi/ctl) library.

```bash
$> ipfs-orchestrator ctl help
$> ipfs-orchestrator -dev ctl StartNetwork Network=test-network
$> ipfs-orchestrator -dev ctl NetworkStats Network=test-network
$> ipfs-orchestrator -dev ctl StopNetwork Network=test-network
```
