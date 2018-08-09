# ptrtar

ptrtar is a tool for creating and extracting tar archives that 
contain pointers to data, instead of the real file data. What the pointer
actually means is delegated to a user supplied command.

The output tar just a regular tar archive, that can be extracted with tar,
but the contents of files are user defined strings like URLS. When ptrtar
extracts the archive it reads this pointer(say, a url) and delegates fetching to another command.

To tell if a tar header file contains a pointer, it has a PAX header 'PTRTAR.sz' = TrueSize

The primary rationale of ptrtar is a directory index format for deduplicated/encrypted
backups, while preserving the unix spirit of simple composable tools. If you create lots
of similar ptrtar archives, the pointer can be the same for the same files, and therefore save space, even
after file encryption. The ptrtar archives themselves can be compressed and encrypted too.

walkthrough, gpg encrypted backups into google cloud:

```
# consider the example scripts, uploadencryptedtogoogle and downloadencryptedfromgoogle
# upload reads the file contents from stdin. encrypts it, uploads the data
# and prints a url to stdout.

gpgid=ac@acha.ninja
bucket=gs://mytopsecretbucket/

$ echo hello | ./example/uploadencryptedtogoogle.sh $bucket $gpgid
gs://mytopsecretbucket/f572d396fae9206628714fb2ce00f72e94f2258f

# download takes a url from stdin, decrypts it, and prints the file contents.

$ echo gs://mytopsecretbucket/f572d396fae9206628714fb2ce00f72e94f2258f | ./example/downloadfromgoogle.sh
hello

# now combining it with ptrtar create, we can encrypt our files and securely upload
# them to google.

$ ptrtar create -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $gpgid > files.ptrtar

# remember, a ptrtar archive is just a tar archive, regular tar can extract it, just
# your file data is replaced with your pointers! (tar will also mention some unknown metadata).

# its a bit slow though, that sucks.
# we can use a cache file, that remembers the url of files we uploaded, 
# the cache is just an sqlite3 database remembering full path, modified time, and file size
# so future archives don't need to reupload if the file didn't change.

# first backup is slow.

$ ptrtar create -cache /tmp/ptrtar.cache -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $ gpgid > backup1.ptrtar

# second backup is hella fast, hell yeah!
# only new files and changed files are reencrypted and uploaded
# Yet the backup is a full snapshot, the cache
# gave us extremely efficient snapshots!

$ ptrtar create -cache /tmp/ptrtar.cache -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $ gpgid > backup2.ptrtar

# if we want to to see what pointers, or in this case, google urls a ptrtar contains
# just run list-ptrs

$ ptrtar list-ptrs < backup1.ptrtar 
gs://mytopsecretbucket/f40dc2ae20adb3883d3e0eda2fde0961243f8132
.
.
.
gs://mytopsecretbucket/fe39f64560a2514fc03aeb6e833405b2729ce751

# if we remove backup1.ptrtar, we should also remove
# the unused objects from google cloud storage.
# this is actually quite easy with some bash foo...
# this command lists all objects in the bucket, also
# all objects in the ptrtar, finds the difference, 
# then deletes those we don't need anymore (I love bash sometimes).

$ comm -23 <(gsutil ls | sort) <(ptrtar list-ptrs < backup2.ptrtar | sort) | gsutil rm -I

# we can convert our ptrtar back to a regular tar file to make sure
# we still have our all our data...

$ ptrtar to-tar ./example/downloadencryptedfromgoogle.sh < backup2.ptrtar > files.tar

# Of course, we still need a place to put our ptrtar files, they have deduplicated and
# speed up our backup process, but we still should encrypt the ptrtar files themselves
# and upload them along side the data objects.

$ gzip < backup2.ptrtar | gpg --encrypt -r $gpgid | gsutil cp - $bucket/backups.ptrtar.gz.enc

# One difficulty is if you delete your files, you must invalidate your backup cache files, 
# or else your new archives may point to data you already deleted.

$ rm /tmp/ptrtar.cache

# In this sense, ptrtar is a very low level tool, but I'm sure higher level tools and scripts
# could use it, or it's code as a building block.
```


# TODO

## definitely:

- worker pool running the upload command.
- embed a "multihash" with each ptr so the contents can
  be verified.

## likely:

- A spec+header for nesting ptrtars
  - option create -nest
  - make to-tar understand nesting.
  - make ptr-list understand nesting.

## possible:


- ptrtar fuse that lazily downloads pointers to a cache.
- example of ptrtar pointing into ipfs or bittorrent.
- contrib scripts, 'garbage collection'
- contrib script doing backups into git annex
- A separate content chunking tool that can be composed with ptrtar.
  ptrtar is general enough that you could content chunk the ptrtar archive
  as well as before you generate the pointers themselves.
