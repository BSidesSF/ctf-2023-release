# Summary

Files is a binary exploitation challenge which allows a user to upload and download text files from a database. The `upload` command contains a format string vulnerability which can be used to leak a memory address of a user controlled buffer. The `download` command is vulnerable to a buffer overflow vulnerability since the downloaded file is placed into a buffer with a size of 1000 bytes but the maximum size of a file that can be uploaded is 2000 bytes. These two vulnerabilities together allow for remote code execution by first leaking a memory address with the `upload` command and uploading our payload, then downloading our payload and returning to the leaked memory address which stores our shellcode.

