BUILD?=$(shell basename `pwd`)-build

all:
	docker build . -t ${BUILD} -f Dockerfile.build
	docker run --privileged --rm -v ${PWD}:/src --env CFLAGS="${CFLAGS}" --env UID=$(shell id -u) --env GID=$(shell id -g) -ti ${BUILD}

indocker:
	cd src && make
	mv src/overflowme .

clean:
	rm -f overflowme
	cd src && make clean
