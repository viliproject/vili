NAME := quay.io/airware/vilibase
VERSION := $(shell date "+%Y%m%d-%H%M%S")

.PHONY: build publish

all: build

build:
	docker build -t ${NAME}:${VERSION} .

publish: build
	docker push ${NAME}:${VERSION}
