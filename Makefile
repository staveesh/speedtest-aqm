GOCMD=go
GOBUILD=${GOCMD} build

CMD_DIR=./cmd
BIN_DIR=./bin
NETRICS?=false

ifeq ($(NETRICS), true)
	CMD_NAME=netrics
	BIN_NAME=netrics-traceneck
	DOCKER_TARGET=netrics
	DOCKER_TAG=netrics
else
	CMD_NAME=main
	BIN_NAME=traceneck
	DOCKER_TARGET=main
	DOCKER_TAG=latest
endif

CMD=${CMD_DIR}/${CMD_NAME}
BIN=${BIN_DIR}/${BIN_NAME}

GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)
GIT_DIRTY=$(shell git diff --quiet --ignore-submodules HEAD 2>/dev/null || echo '!')
ifeq ($(GIT_COMMIT),)
	VERSION?=dev
else
	VERSION?=git (${GIT_COMMIT}${GIT_DIRTY})
endif

MODULE=$(shell go list -m)
LD_FLAGS=-X '${MODULE}/internal/config.NAME=${BIN_NAME}' -X '${MODULE}/internal/config.VERSION=${VERSION}'

all: build

deps: go.mod go.sum
	${GOCMD} mod download

clean:
	${GOCMD} clean
	rm -rf ${BIN_DIR}

build: ${CMD}
	${GOBUILD} -o ${BIN} -ldflags="${LD_FLAGS}" ${CMD}

release: ${CMD}
	${GOBUILD} -o ${BIN} -ldflags="${LD_FLAGS} -s -w" -trimpath ${CMD}

setcap: ${BIN}
	sudo setcap cap_net_raw,cap_net_admin=eip ${BIN}

.PHONY: docker
docker:
	docker build \
	--build-arg NAME=${BIN_NAME} \
	--build-arg NETRICS=${NETRICS} \
	--build-arg VERSION="${VERSION}" \
	-t traceneck:${DOCKER_TAG} .
