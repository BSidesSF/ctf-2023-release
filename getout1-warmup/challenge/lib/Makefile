override CFLAGS += -Wall -fPIC
LIBS=libgetout.o util.o packet.o

all: ${LIBS}
	${CC} -shared ${CFLAGS} -o ../libgetout.so ${LIBS} -lz

clean:
	rm -f *.o test
