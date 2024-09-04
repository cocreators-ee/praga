#!/usr/bin/env bash

# Runs praga with the nginx user, giving the nginx user access to the unix socket for connections
# Tries to kill the container if praga shuts down e.g. due to misconfiguration

function run_praga {
  su nginx -s /bin/bash -c 'praga --config=/etc/praga.yaml' > /dev/console 2>&1
  # Try to kill the container if praga crashes
  kill "$(find /proc -wholename "/proc/*/task/*/exe" 2>/dev/null | cut -d/ -f3)"
}

mkdir /run/praga
chown nginx /run/praga

run_praga &
