#!/bin/bash
CMD=$1
IFACE=$2
BW=$3  # Bandwidth in kbps
LATENCY=$4  # Delay in ms
LOSS=$5  # Packet loss percentage
AQM_METHOD=$6  # Can be "fq_codel", "cake", "pie", or "no_aqm"

TC=/sbin/tc

start() {
    if [[ -z "$BW" || -z "$LATENCY" || -z "$LOSS" ]]; then
        echo "Usage: $0 start <interface> <bandwidth_kbps> <latency_ms> <loss_percentage> <aqm_method>"
        exit 1
    fi

    echo "Applying traffic shaping with AQM ($AQM_METHOD) to $IFACE..."

    # Clear any existing qdisc
    $TC qdisc del dev $IFACE root 2>/dev/null

    # Add HTB root qdisc and set bandwidth
    $TC qdisc add dev $IFACE root handle 1: htb default 11
    $TC class add dev $IFACE parent 1: classid 1:11 htb rate ${BW}kbit ceil $((BW * 12 / 10))kbit burst $((BW * 18 / 1000))k cburst $((BW * 18 / 1000))k

    # Add netem for delay and packet loss
    $TC qdisc add dev $IFACE parent 1:11 handle 10: netem delay ${LATENCY}ms loss ${LOSS}%

    # Apply AQM method or fallback to tail drop if "no_aqm"
    if [[ "$AQM_METHOD" == "no_aqm" ]]; then
        $TC qdisc add dev $IFACE parent 10: handle 20: pfifo limit 100
        echo "Traffic shaping applied with tail drop (pfifo) on $IFACE"
    else
        $TC qdisc add dev $IFACE parent 10: handle 20: ${AQM_METHOD}
        echo "Traffic shaping applied with $AQM_METHOD on $IFACE"
    fi
}

stop() {
    echo "Removing traffic shaping from $IFACE..."
    $TC qdisc del dev $IFACE root 2>/dev/null
}

show() {
    echo "Current traffic control settings for $IFACE:"
    $TC qdisc show dev $IFACE
    $TC class show dev $IFACE
    $TC filter show dev $IFACE

    # Check if AQM is applied
    AQM_CHECK=$($TC -s qdisc show dev $IFACE | grep -E "fq_codel|cake|pie|sfq|codel")
    if [[ -n "$AQM_CHECK" ]]; then
        echo "✅ AQM ($AQM_METHOD) is active on $IFACE."
    elif [[ "$AQM_METHOD" == "no_aqm" ]]; then
        echo "✅ Tail drop (pfifo) is applied instead of AQM on $IFACE."
    else
        echo "❌ AQM is NOT applied!"
    fi
}

case "$CMD" in
    start)
        start
        show  # Automatically show status after applying AQM
        ;;
    stop)
        stop
        ;;
    show)
        show
        ;;
    *)
        echo "Usage: $0 {start|stop|show} <interface> <bandwidth_kbps> <latency_ms> <loss_percentage> <aqm_method>"
        ;;
esac