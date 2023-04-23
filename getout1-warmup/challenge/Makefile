override CFLAGS += -Wall -fno-stack-protector

BUILD?=$(shell basename `pwd`)-build

GETOUTRPC_LIBS=server/getoutrpc.o server/services.o server/server.o libgetout.so
RPCPING_LIBS=services/ping/rpcping.o libgetout.so
GETTOKEN_LIBS=services/gettoken/rpcgettoken.o libgetout.so
APPLY_LIBS=services/apply/apply.o libgetout.so
SHUTDOWN_LIBS=level3-finale/shutdown.o

all:
	docker build . -t ${BUILD} -f Dockerfile.build
	docker run --privileged --rm -v ${PWD}:/src --env CFLAGS="${CFLAGS}" --env UID=$(shell id -u) --env GID=$(shell id -g) -ti ${BUILD}

indocker: libgetout.so getoutrpc rpcping apply shutdown gettoken

libgetout.so:
	cd lib && make

rpcping: ${RPCPING_LIBS}
	${CC} ${CFLAGS} -o rpcping ${RPCPING_LIBS}
	strip rpcping

gettoken: ${GETTOKEN_LIBS}
	${CC} ${CFLAGS} -o gettoken ${GETTOKEN_LIBS}
	strip gettoken

apply: ${APPLY_LIBS}
	${CC} ${CFLAGS} -o apply ${APPLY_LIBS} -lssl -lcrypto -no-pie
	strip apply

# The main RPC server
getoutrpc: ${GETOUTRPC_LIBS}
	${CC} ${CFLAGS} -o getoutrpc ${GETOUTRPC_LIBS}
	strip getoutrpc

shutdown: ${SHUTDOWN_LIBS}
	${CC} ${CFLAGS} -o shutdown ${SHUTDOWN_LIBS}

clean:
	rm -f *.o libgetout.so getoutrpc test shutdown apply gettoken register rpcping
	rm -f server/*.o
	rm -f services/*/*.o

	cd lib && make clean
