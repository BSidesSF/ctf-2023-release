FROM ubuntu:22.04

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install socat -y && \
    apt-get install sqlite3 -y

WORKDIR /home/ctf

RUN useradd -m ctf -u2000

COPY files ./files
COPY flag.txt ./flag.txt

RUN chown -R root:ctf /home/ctf
RUN chmod 777 .

EXPOSE 6464

USER 2000

CMD ["socat", "TCP4-LISTEN:6464,fork", "EXEC:/home/ctf/files"]
