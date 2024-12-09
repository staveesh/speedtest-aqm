# syntax=docker/dockerfile:1

ARG GO_IMAGE=golang:1.23
ARG DIST_IMAGE=debian:12-slim
ARG OOKLA_VERSION=1.2.0
ARG NDT7_VERSION=latest
ARG DEBIAN_FRONTEND=noninteractive

ARG CLIENT_DIR=/client
ARG BUILD_DIR=/bf
ARG BIN_TAR=bin.tgz
ARG VERSION

FROM ${GO_IMAGE} AS client

ARG TARGETPLATFORM
ARG OOKLA_VERSION
ARG NDT7_VERSION
ARG CLIENT_DIR

WORKDIR ${CLIENT_DIR}

RUN <<GET-OOKLA
#!/bin/bash -x
set -euo pipefail
suffix=""
if [ "$TARGETPLATFORM" = "linux/amd64" ] || [ "$(uname -m)" = "x86_64" ]; then
  suffix="x86_64"
elif [ "$TARGETPLATFORM" = "linux/arm64" ] || [ "$(uname -m)" = "aarch64" ]; then
  suffix="aarch64"
else
  echo "Unsupported platform"
  exit 1
fi

wget -qO- https://install.speedtest.net/app/cli/ookla-speedtest-${OOKLA_VERSION}-linux-${suffix}.tgz |
  tar xz speedtest
GET-OOKLA

RUN <<GET-NDT
#!/bin/bash -x
set -euo pipefail
export GOBIN=$(pwd)
go install -ldflags="-s -w" -trimpath -v github.com/m-lab/ndt7-client-go/cmd/ndt7-client@${NDT7_VERSION}
GET-NDT

FROM ${GO_IMAGE} AS build-base

ARG DEBIAN_FRONTEND
ARG BUILD_DIR

WORKDIR ${BUILD_DIR}

RUN <<APT-GET
#!/bin/bash -x
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends libpcap-dev
APT-GET

COPY go.* .
RUN go mod download

COPY . .

FROM build-base AS bin

ARG DEBIAN_FRONTEND
ARG BIN_TAR
ARG VERSION

RUN <<APT-GET
#!/bin/bash -x
set -euo pipefail
dpkg --add-architecture arm64
apt-get update
apt-get install -y --no-install-recommends gcc-aarch64-linux-gnu libpcap-dev:arm64
APT-GET

ARG CGO_ENABLED=1

RUN <<BUILD
#!/bin/bash -x
set -euo pipefail

make release
tar cvzf amd64.tgz bin/
rm -rf bin/
BUILD

ARG CC=aarch64-linux-gnu-gcc
ARG GOARCH=arm64

RUN <<BUILD
#!/bin/bash -x
set -euo pipefail

make release
tar cvzf arm64.tgz bin/
rm -rf bin/
BUILD

RUN tar cvzf ${BIN_TAR} *.tgz

FROM build-base AS release

ARG VERSION

RUN make release

FROM ${DIST_IMAGE} AS dist

ARG DEBIAN_FRONTEND
ARG CLIENT_DIR

RUN <<APT-GET
#!/bin/bash -x
set -euo pipefail
apt-get update
apt-get install -y --no-install-recommends libpcap-dev ca-certificates tzdata
rm -rf /var/lib/apt/lists/*
APT-GET

COPY --from=client ${CLIENT_DIR}/ndt7-client ${CLIENT_DIR}/speedtest /usr/local/bin

ENTRYPOINT ["traceneck"]

FROM dist AS buildx

ARG TARGETARCH
ARG BIN_TAR

COPY ${BIN_TAR} /tmp

RUN <<INSTALL
#!/bin/bash -x
set -euo pipefail
tar xvzOf /tmp/${BIN_TAR} ${TARGETARCH}.tgz |
  tar xvz -C /usr/bin --strip-components=1 bin/traceneck
rm -rf /tmp/*
INSTALL

FROM dist AS main

ARG BUILD_DIR

COPY --from=release ${BUILD_DIR}/bin/* /usr/local/bin/
