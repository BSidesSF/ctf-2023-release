*(If you're familiar with ROP already, go ahead and skip to [Level 1](#level1)!)*

The basis of ROP is that you overwrite the call stack with your own call stack.
Each element on the stack is either a `pop`, which moves the next thing on the
stack into a register, or it's a function. If it's a function, the CPU
"returns" directly to the top of the function, and the function acts as if it's
being called.

To begin, let's just write a function that exits cleanly (ie, with an error
code of 0). That requires two things:

* First, we want to set the `rdi` register to zero (0)
* Second, we want to return to the `exit(code)` function

Why set `rdi` to 0? `rdi` is the first parameter when a function is called (see
[x64 calling conventions](https://en.wikipedia.org/wiki/X86_calling_conventions#System_V_AMD64_ABI),
or mouseover one of the functions in the sidebar), and we want the parameter to
`exit` to be 0 for a clean exit. Then we want to jump straight to `exit`!

To do those two things:

* Click on `pop rdi / ret` at the left - you'll be prompted for the value you want to pop into `rdi` - enter 0!
* Click on `exit(code)` - that means when the `pop / ret` completes, it'll "return" to the top of exit.

You should end up with a stack that looks something like (your addresses will
likely be different):

```
  0000371300000000  0x13370000  pop rdi / ret ; <-- returns to a "pop" instruction
  0000000000000000  0x0         Constant (consumed by pop): 0 ; <-- this is what gets popped into rdi
  7011400000000000  0x401170    exit(code)    ; <-- since the 0 was popped, this is the next thing for "ret" to use
```

When you execute it, you should see the output:

```
  Process exited cleanly with exit code 0
```

Try different functions, and play around!

By the way, if this was a real overflow, you'd see a payload that's just the
stack (ie, left column), like:

```
  echo -ne "\x00\x00\x37\x13\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x70\x11\x40\x00\x00\x00\x00\x00" | ncat <target>
```
