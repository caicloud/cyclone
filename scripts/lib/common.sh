#!/bin/bash
#
#This script provides common script functions for the hacks

# Asks golang what it thinks the host platform is.  The go tool chain does some
# slightly different things when the target platform matches the host platform.
function os::build::host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}
readonly -f os::build::host_platform

function disable-selinux() {
  if selinuxenabled && [ "$(getenforce)" = "Enforcing" ]; then
    echo "Temporarily disabling selinux enforcement"
    sudo setenforce 0
    SELINUX_DISABLED=1
  fi
}
