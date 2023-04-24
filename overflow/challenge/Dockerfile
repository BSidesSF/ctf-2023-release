FROM i386/debian:bullseye-slim

RUN apt update -y && \
apt install socat -y && \
apt clean -y


RUN useradd -m ctf -u1000
RUN useradd -m target -u1001

ADD bashrc /home/ctf/.bashrc
RUN chmod 0755 /home/ctf/.bashrc

ADD flag.txt /home/ctf
RUN chmod 0600 /home/ctf/flag.txt
RUN chown 1001:1001 /home/ctf/flag.txt

ADD src/overflowme.c /home/ctf
RUN chmod 0644 /home/ctf/overflowme.c
RUN chown target:target /home/ctf/overflowme.c

ADD overflowme /home/ctf
RUN chown 1001:1001 /home/ctf/overflowme
RUN chmod 0755 /home/ctf/overflowme
RUN chmod +s /home/ctf/overflowme

USER 1000

ENTRYPOINT ["socat", "TCP-LISTEN:4445,reuseaddr,fork", "EXEC:'/bin/bash -li',pty,stderr,setsid,sane"]
