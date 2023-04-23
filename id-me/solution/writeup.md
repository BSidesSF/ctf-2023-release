I designed `id-me` to be a fairly straight forward "identify this file"
challenge. The user is given four files, and they are tasked with reading part
of the flag from each of them.

I'd personally use the `file` command on Linux:

```sh
$ file *
file1: ASCII text
file2: JPEG image data [...]
file3: PDF document, version 1.4
file4: Zip archive data, at least v2.0 to extract, compression method=deflate
```

But lots of other ways exist, including simply opening them in the Chrome
browser.
