# Traceneck

Bottleneck Link Locator in Access networks

## Prerequisites

- Linux (*Windows and MacOS are untested*)
- [libpcap](https://github.com/the-tcpdump-group/libpcap) library
- [ndt7-client](https://github.com/m-lab/ndt7-client-go) or [ookla](https://www.speedtest.net/apps/cli) speedtest client
- [tshark](https://tshark.dev/setup/install/) (optional)

[Download Binary](https://github.com/internet-equity/traceneck/releases/latest) |
[Docker Image](https://github.com/internet-equity/traceneck/pkgs/container/traceneck)

## Build from Source

Install [go](https://go.dev/dl/), [make](https://www.gnu.org/software/make/) (recommended) and run:

```sh
make build setcap

# OR release:
# make release setcap

./bin/traceneck
```

OR use [Docker](https://docs.docker.com/engine/install/):

```sh
make docker

docker run --rm -v ./data:/data traceneck:latest
```

## Options

```sh
Usage: traceneck [OPTIONS]

Options:
  -I, --interface string   Interface (default "enp0s31f6")
  -t, --tool string        Speedtest tool to use: ndt or ookla (default "ndt")
  -s, --server string      IP address for custom server. Optional. If not provided, will use default server.
  -n, --no-ping            Skip pings
  -p, --ping-type string   Ping packet type: icmp or udp (default "icmp")
  -m, --max-ttl int        Maximum TTL until which to send pings (default 5)
  -d, --direct-hop int     Hop to ping directly by icmp echo [0 to skip] (default 1)
  -T, --tshark             Use TShark
  -i, --idle int           Post speedtest idle time (in secs) (default 10)
  -o, --out-path string    Output path [path with trailing slash for directory, file path for tar archive, "-" for stdout] (default "data/")
  -r, --terse-metadata     Terse rtt metadata
  -q, --quiet              Minimize logging
  -y, --yes                Do not prompt for confirmation
  -h, --help               Show this help
  -v, --version            Show version
```
