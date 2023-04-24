FROM debian:bullseye-slim

# Create a non-root user
RUN useradd -r ctf -u 1000

RUN apt update
RUN apt install -y ncat socat uuid-runtime procps

RUN mkdir /ctf
WORKDIR /ctf

# General stuff
COPY getoutrpc /ctf
RUN chmod 0755 /ctf/getoutrpc
COPY rpcservices /ctf
RUN chmod 0644 /ctf/rpcservices
COPY libgetout.so /lib
RUN chmod 0755 /lib/libgetout.so

# Level 1
COPY rpcping /ctf
RUN chmod 0755 /ctf/rpcping
COPY level1-flag.txt /ctf
RUN chmod 0444 /ctf/level1-flag.txt
COPY level1-narrative.txt /ctf
RUN chmod 0444 /ctf/level1-narrative.txt

# Level 2
COPY gettoken /ctf
RUN chmod 0755 /ctf/gettoken
COPY level2-flag.txt /ctf
RUN chmod 0444 /ctf/level2-flag.txt
COPY level2-narrative.txt /ctf
RUN chmod 0444 /ctf/level2-narrative.txt

# Level 3
COPY apply /ctf
RUN chmod 0755 /ctf/apply
COPY shutdown /bin/shutdown
RUN chmod 0755 /bin/shutdown
COPY level3-flag.txt /ctf/level3-flag.txt
RUN chmod 0444 /ctf/level3-flag.txt
COPY level3-narrative.txt /ctf/level3-narrative.txt
RUN chmod 0444 /ctf/level3-narrative.txt
COPY level3-token.txt /ctf/level3-token.txt
RUN chmod 0444 /ctf/level3-token.txt

# Drop privileges after uploading the files
USER 1000

CMD ["/ctf/getoutrpc", "1337"]
