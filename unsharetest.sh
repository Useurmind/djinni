#!/bin/bash

set -e

unshare -rm -U --map-user=$(id -u) --map-group=$(id -g) bash -c '
set -e
id
mount -t overlay overlay -o lowerdir=overlaymount/lower,upperdir=overlaymount/upper,workdir=overlaymount/work overlaymount/m
su jochen
id
podman run -it --entrypoint bash -v overlaymount/m:/mymount  ubuntu
'