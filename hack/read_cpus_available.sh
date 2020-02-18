#!/bin/bash

set -e

CPUS_AVAILABLE=1

case "$(uname -s)" in
Darwin)
    CPUS_AVAILABLE=$(sysctl -n machdep.cpu.core_count)
    ;;
Linux)
    CFS_QUOTA=$(cat /sys/fs/cgroup/cpu/cpu.cfs_quota_us)
    if [ $CFS_QUOTA -ge 100000 ]; then
    CPUS_AVAILABLE=$(expr ${CFS_QUOTA} / 100 / 1000)
    fi
    ;;
*)
    # Unsupported host OS. Must be Linux or Mac OS X.
    ;;
esac

echo ${CPUS_AVAILABLE}