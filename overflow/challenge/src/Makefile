all: overflowme

overflowme:
	gcc -fno-stack-protector -z execstack -no-pie -m32 -z norelro -g -o overflowme overflowme.c

clean:
	rm -f overflowme *.o
