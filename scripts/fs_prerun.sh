#!/usr/bin/env bash

set -x

# Set up partition and mount
# fs_prerun [fs type] [device] [mount point]

sudo -n blkdiscard $2
sudo -n wipefs -a $2
sudo -n mkfs -t $1 $2
sudo -n mkdir -p $3
sudo -n mount $2 $3
sudo -n chown -R $(id -u) $3
