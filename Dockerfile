ARG GO_IMAGE=golang:latest
ARG DIST_IMAGE=debian:12-slim
ARG OOKLA_VERSION=1.2.0
ARG NDT7_VERSION=latest
ARG DEBIAN_FRONTEND=noninteractive

ARG CLIENT_DIR=/client
ARG BUILD_DIR=/bf
ARG BIN_TAR=bin.tgz
ARG NETRICS
ARG VERSION

FROM ${GO_IMAGE} AS client

ARG TARGETPLATFORM
ARG OOKLA_VERSION
ARG NDT7_VERSION
ARG CLIENT_DIR

WORKDIR ${CLIENT_DIR}

RUN \
  suffix=""; \
  if [ "$TARGETPLATFORM" = "linux/amd64" ] || [ "$(uname -m)" = "x86_64" ]; then \
  suffix="x86_64"; \
  elif [ "$TARGETPLATFORM" = "linux/arm64" ] || [ "$(uname -m)" = "aarch64" ]; then \
  suffix="aarch64"; \
  else echo "Unsupported platform"; exit 1; \
  fi && \
  wget -qO- https://install.speedtest.net/app/cli/ookla-speedtest-${OOKLA_VERSION}-linux-${suffix}.tgz \
  | tar xz speedtest

RUN \
  GOBIN=$(pwd) go install -ldflags="-s -w" -trimpath \
  -v github.com/m-lab/ndt7-client-go/cmd/ndt7-client@${NDT7_VERSION}

FROM ${GO_IMAGE} AS build-base

ARG DEBIAN_FRONTEND
ARG BUILD_DIR

WORKDIR ${BUILD_DIR}

RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends libpcap-dev

COPY go.* .
RUN go mod download

COPY . .

FROM build-base AS bin

ARG DEBIAN_FRONTEND
ARG BIN_TAR
ARG VERSION

RUN \
  dpkg --add-architecture arm64 && \
  apt-get update && \
  apt-get install -y --no-install-recommends \
  gcc-aarch64-linux-gnu libpcap-dev:arm64 

ARG CGO_ENABLED=1

RUN \
  make release && \
  NETRICS=true make release && \
  tar cvzf amd64.tgz bin/ && \
  rm -rf bin

ARG CC=aarch64-linux-gnu-gcc
ARG GOARCH=arm64

RUN \
  make release && \
  NETRICS=true make release && \
  tar cvzf arm64.tgz bin/ && \
  rm -rf bin

RUN tar cvzf ${BIN_TAR} *.tgz

FROM build-base AS release

ARG NETRICS
ARG VERSION

RUN make release

FROM ${DIST_IMAGE} AS dist

ARG DEBIAN_FRONTEND
ARG CLIENT_DIR
ARG NETRICS

RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends \
  libpcap-dev ca-certificates tzdata && \
  rm -rf /var/lib/apt/lists/*

COPY --from=client ${CLIENT_DIR}/ndt7-client /usr/bin
COPY --from=client ${CLIENT_DIR}/speedtest /usr/bin

RUN \
  echo "#!/bin/bash\n" >/entrypoint.sh && \
  echo "/usr/bin/$([ "$NETRICS" = "true" ] && echo "netrics-" || :)traceneck \"\$@\"" \
  >>/entrypoint.sh && chmod +x /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

FROM dist AS buildx

ARG TARGETARCH
ARG BIN_TAR

COPY ${BIN_TAR} /tmp

RUN \
  tar xvzOf /tmp/${BIN_TAR} ${TARGETARCH}.tgz | \
  tar xvz -C /usr/bin --strip-components=1 \
  bin/$([ "$NETRICS" = "true" ] && echo "netrics-" || :)traceneck && \
  rm -rf /tmp/*

FROM dist AS main

ARG BUILD_DIR

COPY --from=release ${BUILD_DIR}/bin/* /usr/bin/
