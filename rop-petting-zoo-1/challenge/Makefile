BUILD?=$(shell basename `pwd`)-build

all:
	docker build . -t ${BUILD} -f Dockerfile.build
	docker run --privileged --rm -v ${PWD}:/src --env CFLAGS="${CFLAGS}" --env UID=$(shell id -u) --env GID=$(shell id -g) -ti ${BUILD}

indocker:
	cd target && make

clean:
	cd target && make clean
