GO=go

main: cosypher

cosypher: cosypher.go
	GOPATH=`pwd` $(GO) build cosypher.go

clean:
	rm -f cosypher
	rm -f *.o
	rm -f *~
	rm -f \#*
