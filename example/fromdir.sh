#! /bin/sh

# This is the inverse of todir.sh, see its file for more details.
set -u
set -e

archivedir=$1
fhash=$(cat)
cat $archivedir/$fhash
