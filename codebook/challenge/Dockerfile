FROM gcr.io/bsides-sf-ctf-2023/base:latest

RUN apt-get install -y perl imagemagick libdigest-sha-perl

WORKDIR /home/ctf

COPY --chown=ctf:ctf . /home/ctf/

USER 2000

CMD ["/home/ctf/run.sh"]
