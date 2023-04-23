#!/usr/bin/env python3

import os
import subprocess
import sys
import shlex
import time


pname = "/home/ctf/shurdles"
if len(sys.argv) > 1:
    pname = sys.argv[1]
cmd = [pname, "foo"]
env = {k: v for k,v in os.environ.items()}
env['HACKERS'] = 'hack\nthe\nplanet'
env['TZ'] = 'America/Los_Angeles'
env['PATH'] = '/opt/shurdles-tools:' + env['PATH']
bashexec = ["exec", "-a", "/shurdles"] + cmd
bashcmd = ["/usr/bin/bash", "-c", "({})".format(shlex.join(bashexec))]
print(repr(bashcmd))

p = "/home/ctf/.cache/shurdles"
with open(p, "w") as fp:
    fp.write('')
t = time.time()
t -= 60*60*24*2
os.utime(p, (t, t))

fd = os.open("/tmp/foo.txt", os.O_RDWR|os.O_CREAT|os.O_TRUNC)
buf = b"A"*1337
os.write(fd, buf)
if fd != 3:
    os.dup2(fd, 3)
dname = '/run/. -- !!'
try:
    os.mkdir(dname)
except FileExistsError:
    pass

rv = subprocess.run(
        bashcmd,
        executable="/usr/bin/bash",
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        env=env,
        pass_fds=(3,),
        cwd=dname)
print(rv.stdout.decode('utf-8'))
sys.exit(rv.returncode)
