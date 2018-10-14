#!/bin/bash

# This script installs protoc for given platform (defaults to linux-x86_64).
# See: https://github.com/protocolbuffers/protobuf/releases

VERSION=3.6.1
PLATFORM=$1
if [ -z "$1" ] ; then
  PLATFORM=linux-x86_64
fi

PROTOC_ZIP=protoc-$VERSION-$PLATFORM.zip
echo Downloading $PROTOC_ZIP...
curl -OL https://github.com/google/protobuf/releases/download/v$VERSION/$PROTOC_ZIP

echo Unzipping $PROTOC_ZIP to /usr/local...
sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc

echo Cleaning up...
rm -f $PROTOC_ZIP
