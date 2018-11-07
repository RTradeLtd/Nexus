#! /bin/bash

export GO111MODULE=on

go get -u github.com/mitchellh/gox
go mod tidy

RELEASE=$(git describe --tags --always)
TARGETS="linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64"

mkdir -p release

gox -output="release/ipfs-orchestrator-$RELEASE-{{.OS}}-{{.Arch}}" \
    -ldflags "-X main.Version=$RELEASE" \
    -osarch="$TARGETS" \
    .
