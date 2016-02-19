SHA := $(shell git rev-parse --short HEAD)

.PHONY: build

all: build

build:
	-rm -rf main build
	docker build -t vilibuilder:${SHA} -f Dockerfile .
	-docker rm -f $(NAME)builder
	docker run -d --name vilibuilder vilibuilder:${SHA} echo 0
	docker cp vilibuilder:/go/src/github.com/airware/vili/main .
	docker cp vilibuilder:/go/src/github.com/airware/vili/public/build .
	docker build -t vili:${SHA} -f Dockerfile.minimal .
	-rm -rf main build

test: lint coverage

lint:
	@echo "todo"

coverage:
	@echo "todo"
