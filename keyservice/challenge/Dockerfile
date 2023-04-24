FROM gcr.io/bsides-sf-ctf-2023/base:latest

COPY keyservice /home/ctf
COPY banner.txt /home/ctf
COPY flag.txt /home/ctf

RUN chmod -R ugo=rX /home/ctf
RUN chown -R root:ctf /home/ctf

EXPOSE 1031

WORKDIR /home/ctf
USER 2000

CMD ["/usr/bin/socat", "TCP4-LISTEN:1031,reuseaddr,fork", "EXEC:/home/ctf/keyservice,pty,stderr,setsid,sane"]
