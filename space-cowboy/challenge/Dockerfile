FROM us-west1-docker.pkg.dev/bsides-sf-ctf-2023/challenges/base:latest
RUN apt-get update -y && \
    apt-get install -y python3-pip python3-dev

COPY ./server /app

WORKDIR /app

RUN pip install -r requirements.txt

RUN chown -R root:ctf /app
RUN chmod -R ug+rwX /app

EXPOSE 8000

USER 2000

ENTRYPOINT [ "python3" ]

CMD [ "server.py" ]
