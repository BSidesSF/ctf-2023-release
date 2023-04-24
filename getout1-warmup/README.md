Quick notes...

To compile: `cd challenge` / `make` - it will use docker (though I've been testing with podman)

To run a specific level in debug mode, provide the port and services file:

```
make clean && CFLAGS=-DDEBUG make && LD_LIBRARY_PATH=. ./getoutrpc (random 1025 50000) ./rpcservices.test
```

To run it in Docker:

```
 make clean && docker build . -t test && docker run -p1337:1337 --rm -ti test
```
