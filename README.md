# 🦑 ipfs-orchestrator

The IPFS private network node orchestration and registry service for
[Temporal](https://github.com/RTradeLtd/Temporal), an easy-to-use interface into
distributed and decentralized storage technologies.

[![GoDoc](https://godoc.org/github.com/RTradeLtd/cmd?status.svg)](https://godoc.org/github.com/RTradeLtd/cmd)
[![Build Status](https://travis-ci.com/RTradeLtd/cmd.svg?branch=master)](https://travis-ci.com/RTradeLtd/cmd)
[![codecov](https://codecov.io/gh/RTradeLtd/cmd/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/cmd)
[![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/cmd)](https://goreportcard.com/report/github.com/RTradeLtd/cmd)

## Installation and Usage

Coming soon!

## Development

This project requires [Docker CE](https://docs.docker.com/install/#supported-platforms)
and [Go 1.11](https://golang.org/dl/) or above.

To fetch the codebase, clone the repository or use `go get`:

```bash
$> go get github.com/RTradeLtd/ipfs-orchestrator
```

Dependencies can be installed using the provided Makefile:

```bash
$> make   # installs dependencies and build binary
```

### Testing

To execute the tests, make sure the Docker daemon is running and run:

```bash
$> make test
```

You can remove leftover assets using `make clean`.
