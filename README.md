# go-dropbox-uploader (WIP)

```bash
$ dropbox_uploader --help
If the file is bigger than 150Mb the file is uploaded using small chunks (default 50Mb);
                in this case a . (dot) is printed for every chunk successfully uploaded and a * (star) if an 
                error occurs (the upload is retried for a maximum of three times). Only if the file is smaller than 150Mb,
                the standard upload API is used, and if the -p option is specified the default curl progress bar is displayed
                 during the upload process. The local file/dir parameter supports wildcards expansion.

Usage:
  upload [flags]

Flags:
  -d, --destination string   dropbox file destination path.
  -h, --help                 help for upload
  -m, --mode string          Selects what to do if the file already exists, default overwite. (default "overwrite")
  -s, --source string        local file source path.
  -t, --token string         dropbox api access token.
```
