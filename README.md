# ptrtar

ptrtar is a tool for creating and extracting tar archives that 
contain pointers to data, instead of the real file data. What the pointer
actually means is delegated to a user supplied command.

The output tar just a regular tar archive, that can be extracted with tar,
but the contents of files are user defined strings like URLS. When ptrtar
extracts the archive it reads this url and delegates fetching to another command.

To tell if a tar header file contains a pointer, it has a PAX header 'PTRTAR.sz' = Size

The primary rationale of ptrtar is a directory index format for deduplicated/encrypted
backups, while preserving the unix spirit of simple composable tools. If you create lots
of similar ptrtar archives, ptr key can be the same for the same files, save space, even
after file encryption. The ptrtar archives themselves can be and compressed encrypted too.

examples:

```
# imagine you have two scripts, uploadtoS3 and downloadfromS3.
# upload reads the file contents from stdin and prints a url to stdout.
# download takes a url from stdin and prints the file contents.

ptrtar create -dir DIR ./uploadtoS3 > files.ptrtar
ptrtar to-tar ./downloadfromS3 < files.ptrtar > files.tar

# ptrtar -create also supports a caching strategy to avoid rerunning the subcommand
# for every file, it makes a dramatic difference.

ptrtar create -cache cache.sqlite  ...

# print all pointers in an archive, this is useful if we want to gather unused data
# from s3 (or wherever) and delete it. We could list an s3 bucket, compare all the files, then delete
# ones not used by our archives.
ptrtar print-ptrs < files.ptrtar
```



# TODO

- worker pool running the upload command.
- to-tar command

Possibilities:

- ptrtar fuse that lazily downloads pointers to a cache.
  - Add a PTRTAR.SZ = TrueSize header
- example of ptrtar pointing into ipfs or bittorrent.
- contrib scripts, 'garbage collection'
- contrib script doing backups into git annex
- A separate content chunking tool that can be composed with ptrtar.
  ptrtar is general enough that you could content chunk the ptrtar archive
  as well as before you generate the pointers themselves.
