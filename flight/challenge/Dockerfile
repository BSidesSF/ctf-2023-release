FROM ubuntu:22.04

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install socat -y && \
    apt-get install sqlite3 -y

RUN dpkg --add-architecture i386 \
    && apt-get update \
    && apt-get install -y libc6:i386 libncurses5:i386 libstdc++6:i386 zlib1g:i386

WORKDIR /home/ctf

COPY flight ./flight
COPY flag.txt ./flag.txt

EXPOSE 6464

USER 2000

ENTRYPOINT ["socat", "TCP4-LISTEN:6464,fork", "EXEC:/home/ctf/flight"]
