FROM gcr.io/bsides-sf-ctf-2023/base:latest

COPY cosypher /home/ctf
COPY cosypher.go /home/ctf
COPY banner.txt /home/ctf
COPY flag.txt /home/ctf

RUN chmod -R ugo=rX /home/ctf
RUN chown -R root:ctf /home/ctf

EXPOSE 31415

WORKDIR /home/ctf
USER 2000

CMD ["/usr/bin/socat", "TCP4-LISTEN:31415,reuseaddr,fork", "EXEC:/home/ctf/cosypher,pty,stderr,setsid,sane"]
