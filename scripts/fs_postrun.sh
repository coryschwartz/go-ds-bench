#!/usr/bin/env bash

set -x

# Cleanup after mounting
# fs_postrun [mountpoint]

sudo -n umount $1
sudo -n rm -r $1
