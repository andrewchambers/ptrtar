# ptrtar

ptrtar is a tool for creating and extracting tar archives that 
contain pointers to data, instead of the real file data. What the pointer
actually means is delegated to a user supplied command.

The output tar just a regular tar archive, that can be extracted with tar,
but the contents of files are user defined strings like URLS. When ptrtar
extracts the archive it reads this url and delegates fetching to another command.

To tell if a tar header file contains a pointer, it has a PAX header 'PTRTAR.sz' = TrueSize

The primary rationale of ptrtar is a directory index format for deduplicated/encrypted
backups, while preserving the unix spirit of simple composable tools. If you create lots
of similar ptrtar archives, ptr key can be the same for the same files, save space, even
after file encryption. The ptrtar archives themselves can be and compressed encrypted too.

walkthrough:

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

# now combining it with ptrtar create, we can encrypt our files them and upload
# them to google.

ptrtar create -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $gpgid > files.ptrtar

# its a bit slow though, that sucks.
# we can use a cache file, that remembers the url of files we uploaded, 
# the cache is just an sqlite3 database remembering full path, modified time, and file size
# so future archives don't need to reupload if the file didn't change.

# first backup is slow.
$ ptrtar create -cache /tmp/ptrtar.cache -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $gpgid | gzip > backup1.ptrtar.gz

# second backup is hella fast, hell yeah!
# only new files and changed files are reencrypted and uploaded
$ ptrtar create -cache /tmp/ptrtar.cache -dir DIR ./example/uploadencryptedtogoogle.sh $bucket $gpgid > gzip > backup1.ptrtar.gz

# if we want to to see what pointers, or in this case, s3 urls a ptrtar contains
# just run list-ptrs
gunzip < backup1.ptrtar.gz | ptrtar list-ptrs

# we could even see what changed across our archives

diff <(gunzip < backup1.ptrtar.gz | ptrtar list-ptrs) <(gunzip < backup2.ptrtar.gz | ptrtar list-ptrs)

# lets convert our ptr tar back to a regular tar file so we 
# can extract it.
ptrtar to-tar ./example/downloadencryptedfromgoogle.sh < backup1.ptrtar > files.tar


# Of course, we still need a place to put our ptrtar files, they have deduplicated and
# speed up our backup process, but we still should encrypt the ptrtar files themselves
# and upload them along side the data objects.
```


# TODO

- worker pool running the upload command.

Possibilities:

- ptrtar fuse that lazily downloads pointers to a cache.
- example of ptrtar pointing into ipfs or bittorrent.
- contrib scripts, 'garbage collection'
- contrib script doing backups into git annex
- A separate content chunking tool that can be composed with ptrtar.
  ptrtar is general enough that you could content chunk the ptrtar archive
  as well as before you generate the pointers themselves.
