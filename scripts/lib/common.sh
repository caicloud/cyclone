#!/bin/bash
#
#This script provides common script functions for the hacks

function command_exists() {
    command -v "$@" > /dev/null 2>&1
}


function disable_selinux() {
  if selinuxenabled && [ "$(getenforce)" = "Enforcing" ]; then
    echo "Temporarily disabling selinux enforcement"
    sudo setenforce 0
    SELINUX_DISABLED=1
  fi
}

# set default
: ${os:=""}

# perform some very rudimentary platform detection
if command_exists lsb_release; then
    os="$(lsb_release -si)"
fi
if [ -z "$os" ] && [ -r /etc/lsb-release ]; then
    os="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
fi
if [ -z "$os" ] && [ -r /etc/debian_version ]; then
    os='debian'
fi
if [ -z "$os" ] && [ -r /etc/fedora-release ]; then
    os='fedora'
fi
if [ -z "$os" ] && [ -r /etc/oracle-release ]; then
    os='oracleserver'
fi
if [ -z "$os" ] && [ -r /etc/centos-release ]; then
    os='centos'
fi
if [ -z "$os" ] && [ -r /etc/redhat-release ]; then
    os='redhat'
fi
if [ -z "$os" ] && [ -r /etc/photon-release ]; then
    os='photon'
fi
if [ -z "$os" ] && [ -r /etc/os-release ]; then
    os="$(. /etc/os-release && echo "$ID")"
fi
if [ -z "$os" ] && [[ "$(uname -s)" == "Darwin" ]]; then
    os="darwin"
fi

os="$(echo "$os" | cut -d " " -f1 | tr '[:upper:]' '[:lower:]')"

# Special case redhatenterpriseserver
if [ "${os}" = "redhatenterpriseserver" ]; then
        # Set it to redhat, it will be changed to centos below anyways
        lsb_dist='redhat'
fi

# export -f command_exists >/dev/null 2>&1

export OS=$os


# get this file path
echo "$0" | grep -q "$0"
if [ $? -eq 0 ];
then
    cd "$(dirname ${BASH_SOURCE})"
    CUR_FILE=$(pwd)/$(basename ${BASH_SOURCE})
    CUR_DIR=$(dirname ${CUR_FILE})
    cd - > /dev/null
else
    if [ ${0:0:1} = "/" ]; then
        CUR_FILE=$0
    else
        CUR_FILE=$(pwd)/$0
    fi
    CUR_DIR=$(dirname ${CUR_FILE})
fi

# eliminate relative path ，like a/..b/c
CYCLONE_ROOT=$(dirname $(dirname ${CUR_DIR}))
export CYCLONE_ROOT

echo "cyclone root path:" ${CYCLONE_ROOT}