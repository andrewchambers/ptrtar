#!  /usr/bin/env bash

# this example uses gpg and gsutil to store encrypted
# file data into google.

set -u
set -e
set -o pipefail
set -x

ptrname=$(cat)
gsutil cp $ptrname - | gpg --decrypt | gunzip
