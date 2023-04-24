FROM debian:bookworm

RUN apt-get update -y && \
    apt-get install -y python3-pip python3-dev python3-venv tini nginx

RUN mkdir /config
COPY *.pem /config
COPY nginx.conf /config/nginx.conf

COPY app /app
COPY entrypoint.sh /entrypoint.sh

WORKDIR /app

RUN python3 -m venv /app/venv
RUN /app/venv/bin/pip install -r requirements.txt
RUN adduser ctf

RUN chown -R root:ctf /app
RUN chmod -R ugo+rwX /app

EXPOSE 443

ENTRYPOINT ["/usr/bin/tini"]

CMD ["/entrypoint.sh"]
