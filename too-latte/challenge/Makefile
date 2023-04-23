BUILD?=$(shell basename `pwd`)-build

all:
	docker build . -t ${BUILD} -f Dockerfile.build
	docker run --privileged --rm -v ${PWD}:/src --env UID=$(shell id -u) --env GID=$(shell id -g) -ti ${BUILD}
	zip -r ../distfiles/app.zip ./app/webapp/index.jsp ./app/webapp/WEB-INF

clean:
	rm -rf ./app/build
	rm -f ./app/webapp/WEB-INF/lib/*

# This runs inside Docker, customize this part!
indocker:
	java --version
	javac --version
	cd app/ && ant
