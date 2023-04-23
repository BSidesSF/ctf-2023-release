override CFLAGS += -Wall -Wno-main -fno-pie -no-pie -fno-stack-protector

all: target

target: target.o
	${CC} ${CFLAGS} -o target target.o

clean:
	rm -f *.o target
