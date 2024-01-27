COMMIT_HASH = $(shell git rev-parse --short HEAD)
VERSION ?= ${COMMIT_HASH}
TAG ?= latest
OS ?= linux
ARCH ?= $(shell uname -m|sed 's|x86_64|amd64|'|sed 's|aarch64|arm64|')
GOPROXY ?= https://goproxy.cn,direct
GOVERSION ?= 1.20.11
ORG ?= crazytaxii
TARGET_DIR ?= dist

.PHONY: build image lint

build:
	GOOS=${OS} GOARCH=${ARCH} GOPROXY=${GOPROXY} go build -o ${TARGET_DIR}/ -ldflags "-X 'main.version=${VERSION}'" ./cmd/qos-controller

image:
	docker build --build-arg VERSION=${VERSION} \
		--build-arg GOPROXY=${GOPROXY} \
		--build-arg GOVERSION=${GOVERSION} \
		-f ./Dockerfile \
		-t ${ORG}/volume-qos-controller:${TAG} .

lint:
	golangci-lint run
