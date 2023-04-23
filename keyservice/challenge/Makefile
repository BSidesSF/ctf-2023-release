GO=go

main: keyservice

keyservice: keyservice.go
	GOPATH=`pwd` $(GO) build keyservice.go

clean:
	rm -f keyservice
	rm -f *.o
	rm -f *~
	rm -f \#*
