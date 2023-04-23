FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY main.go ./
RUN go build -v -o lastpwned ./...

FROM alpine:latest

RUN apk add --no-cache tini

RUN addgroup -g 1000 ctf && adduser -u 1000 -G ctf -h /ctf -D ctf
RUN chown 1000:1000 /ctf

WORKDIR /ctf

COPY --from=builder /usr/src/app/lastpwned ./lastpwned
COPY ./static ./static
COPY ./templates ./templates
RUN chown -R 1000:1000 /ctf

USER 1000
EXPOSE 3000/tcp

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/ctf/lastpwned"]
