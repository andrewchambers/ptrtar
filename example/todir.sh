#! /bin/sh

# this example takes a directory, and deduplicates files
# written into it via thier sha1sum, then prints the sha1sum
#
#$ echo hello | ./example/todir.sh /tmp/archive
# f572d396fae9206628714fb2ce00f72e94f2258f
#
# echo f572d396fae9206628714fb2ce00f72e94f2258f | ./example/fromdir.sh /tmp/archive/
# hello
#
# We can use this as a pointer function to ptrtar to deduplicate
# our backups into a dir by file hash.
#
# ptrtar c -dir . ./example/todir.sh /tmp/archive  >  /tmp/test.tar
#

set -u
set -e

archivedir=$1
tempfile=$(mktemp -p $archivedir)
fhash=$(tee $tempfile | sha1sum | awk '{print $1}')
mv $tempfile $archivedir/$fhash
echo $fhash