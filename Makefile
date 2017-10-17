SHA := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date +%s)

.PHONY: build

all: build

build:
	docker build -t vili:${SHA} -f Dockerfile .

publish: build
	docker tag vili:${SHA} quay.io/airware/vili:${TIMESTAMP}-${SHA}
	sleep 1
	docker push quay.io/airware/vili:${TIMESTAMP}-${SHA}

test: lint coverage

lint:
	@echo "todo"

coverage:
	@echo "todo"
