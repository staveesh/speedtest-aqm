#!/bin/bash

IFACE="enp88s0"
IFB="ifb0"
DELAY="20ms"
ACTION=$1  # "add" or "del"

add_netem() {
  echo "🔧 Loading ifb module..."
  sudo modprobe ifb

  echo "🔧 Creating $IFB and setting up..."
  sudo ip link add $IFB type ifb 2>/dev/null
  sudo ip link set $IFB up

  echo "📡 Redirecting ingress traffic from $IFACE to $IFB..."
  sudo tc qdisc add dev $IFACE ingress 2>/dev/null
  sudo tc filter add dev $IFACE parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev $IFB

  echo "⏱ Adding netem delay of $DELAY to $IFB..."
  sudo tc qdisc add dev $IFB root netem delay $DELAY

  echo "✅ Netem delay applied via $IFB."
}

del_netem() {
  echo "🧹 Removing netem and redirect setup..."

  sudo tc qdisc del dev $IFACE ingress 2>/dev/null
  sudo tc qdisc del dev $IFB root 2>/dev/null
  sudo ip link set $IFB down 2>/dev/null
  sudo ip link delete $IFB 2>/dev/null

  echo "✅ Netem delay and ifb redirection removed."
}

case "$ACTION" in
  add)
    add_netem
    ;;
  del)
    del_netem
    ;;
  *)
    echo "Usage: $0 {add|del}"
    exit 1
    ;;
esac
