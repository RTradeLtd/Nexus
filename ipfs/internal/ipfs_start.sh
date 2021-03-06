#!/bin/sh

# Modified IPFS node initialization script.
# Mount to /usr/local/bin/start_ipfs
# Source: https://github.com/ipfs/go-ipfs/blob/${IPFS_VERSION}/bin/container_daemon

set -e

# arguments provided through string templates
DISK_MAX=%dGB

# set variables
user=ipfs
repo="$IPFS_PATH"

# set user
if [ "$(id -u)" -eq 0 ]; then
  echo "changing user to $user"
  # ensure folder is writable
  su-exec "$user" test -w "$repo" || chown -R -- "$user" "$repo"
  # restart script with new privileges
  exec su-exec "$user" "$0" "$@"
fi

# check exec, report version
ipfs version

# check for existing repo - otherwise init new one
if [ -e "$repo/config" ]; then
  echo "found IPFS fs-repo at $repo"
else
  ipfs init --profile server
  ipfs config Addresses.API /ip4/0.0.0.0/tcp/5001
  ipfs config Addresses.Gateway /ip4/0.0.0.0/tcp/8080
fi

# set datastore quota
ipfs config Datastore.StorageMax $DISK_MAX

# release locks
ipfs repo fsck

# if the first argument is daemon
if [ "$1" = "daemon" ]; then
  # filter the first argument until
  # https://github.com/ipfs/go-ipfs/pull/3573
  # has been resolved
  shift
else
  # print deprecation warning
  # go-ipfs used to hardcode "ipfs daemon" in it's entrypoint
  # this workaround supports the new syntax so people start setting daemon explicitly
  # when overwriting CMD
  echo "DEPRECATED: arguments have been set but the first argument isn't 'daemon'" >&2
fi

exec ipfs daemon "$@"
