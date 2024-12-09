GOCMD=go
GOBUILD=${GOCMD} build

CMD_DIR=./
BIN_DIR=./bin/

CMD_FILE=main.go
BIN_NAME=traceneck
DOCKER_TAG=latest

CMD=${CMD_DIR}/${CMD_FILE}
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
	${GOBUILD} -o ${BIN} -ldflags="${LD_FLAGS}" ${CMD_DIR}

release: ${CMD}
	${GOBUILD} -o ${BIN} -ldflags="${LD_FLAGS} -s -w" -trimpath ${CMD_DIR}

setcap: ${BIN}
	sudo setcap cap_net_raw,cap_net_admin=eip ${BIN}

.PHONY: docker
docker:
	docker build \
	--build-arg NAME=${BIN_NAME} \
	--build-arg VERSION="${VERSION}" \
	-t traceneck:${DOCKER_TAG} .
