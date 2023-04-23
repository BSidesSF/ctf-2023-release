Welcome to the second tutorial!

In this, we want to print the string "Hello world!" to stdout. Don't forget
that you can mouseover the gadgets to get a description that includes what each
parameter means!

We have a helpful function that returns a pointer to the string "Hello world!"
for you - it's on the left, and is called `get_hello_world()`. You'll want to
call that one first.

By convention, functions return their value in the `rax` register, and functions
want their first parameter to be in the `rdi` register. We provided a gadget
that'll put the value from `rax` into `rdi` - it's called `mov rdi, rax / ret`.
That's the second thing you want.

Once `rdi` is set correctly, all you need to do is call `puts()` and then
return cleanly using `pop rdi` (with the parameter `0`) followed by `exit()`!

To summarize, you want to build a ROP chain that executes:

* `get_hello_world()`
* `mov rdi, rax / ret`
* `puts()`
* `pop rdi / ret`
* `0`
* `exit()`

If successful, you should see the output `Hello world!`.

Good luck!
