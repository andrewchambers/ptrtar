# ptrtar

ptrtar is a tool for creating and extracting tar archives that 
contain pointers to data, instead of the real file data. What the pointer
actually means is delegated to a user suppliedcommand.

The output tar just a regular tar archive, that can be extracted with tar,
but the contents of files are user defined strings like URLS. When ptrtar
extracts the archive it reads this url and delegates fetching to another command.

To tell if a tar header file contains a pointer, it has a PAX header 'PTRTAR.?' = 'y'

The primary rationale of ptrtar is a directory index format for deduplicated and
gpg encrypted backups, while preserving the unix spirit of simple composable tools.

examples:

```
# imagine you have two scripts, uploadtoS3 and downloadfromS3.
# upload reads the file contents from stdin and prints a url to stdout.
# download takes a url from stdin and prints the file contents.

ptrtar create -dir DIR ./uploadtoS3 > files.tar
ptrtar extract ./downloadfromS3 < files.tar

# ptrtar -create also supports a caching strategy to avoid rerunning the subcommand
# for every file.

ptrtar create -create-cache db.ptrtarcache  ...

# print all pointers in an archive, this is useful if we want to gather unused data
# from s3 (or wherever) and delete it. We could list an s3 bucket, compare all the files, then delete
# ones not used by our archives.
ptrtar print-ptrs < files.tar
```



# TODO

- worker pool running the upload command.
- example of ptrtar pointing into ipfs or bittorrent.
- contrib scripts, 'garbage collection'