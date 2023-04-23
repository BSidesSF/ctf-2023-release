FROM golang:1.20-alpine

RUN apk add --no-cache build-base gtk+3.0-dev gobject-introspection-dev tini
RUN mkdir /app
RUN adduser -u 2000 -D -s /bin/sh ctf

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum

ENV CGO_ENABLED=1
RUN go get -a -v -x github.com/gotk3/gotk3@v0.6.1
RUN go build -v -x github.com/gotk3/gotk3/... || true

COPY . /app

RUN mkdir -p /app/out/ui/lib
RUN cp /lib/ld-musl-x86_64.so.1 /app/out/ui/lib
RUN ln -s ./ld-musl-x86_64.so.1 /app/out/ui/lib/libc.musl-x86_64.so.1

RUN go build -v -o /app/out/ui/ui \
  -ldflags "-linkmode external -extldflags '-Wl,-rpath -Wl,\$ORIGIN/lib -Wl,--dynamic-linker=./lib/ld-musl-x86_64.so.1'" \
  ./ui/...
RUN cd /app/out/ui; ldd ./ui | \
  sed -n 's/.* \(\/[A-Za-z0-9_.\/-]\+\) .*/\1/p' | \
  xargs -i cp -L {} /app/out/ui/lib
RUN tar c -C /app/out/ui -f /app/out/ui.tar '.'

RUN go build -v -o /app/out/server ./server/...
RUN ln -s /app/index.html /app/out/

RUN chown -R ctf:ctf /app

USER 2000

ENTRYPOINT ["/sbin/tini"]
CMD ["/app/out/server"]
