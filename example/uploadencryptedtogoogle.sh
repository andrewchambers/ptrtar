#!  /usr/bin/env bash

# this example uses gpg and gsutil to store encrypted
# file data into google.

set -u
set -e
set -o pipefail
set -x

bucketurl=$1
gpgid=$2
tmpname=$(uuidgen)
datafifo=$(mktemp -u)

mkfifo $datafifo
gzip -9 < $datafifo | gpg --encrypt -r $gpgid  | gsutil cp - $bucketurl/$tmpname > /dev/null &
waitid=$!
sha1name=$(tee $datafifo | sha1sum | awk '{ print $1 }')

wait $waitid
rm $datafifo

gsutil mv $bucketurl/$tmpname $bucketurl/$sha1name > /dev/null
echo $bucketurl/$sha1name
