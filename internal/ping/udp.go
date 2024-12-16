package ping

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

const startingPort = 1024

func handleTimeExceededUDP(replyIP net.IP, recvTime time.Time, msg *icmp.Message) {
	msgBody, ok := msg.Body.(*icmp.TimeExceeded)
	if !ok {
		return
	}

	ipHeaderLen := 40
	if (msgBody.Data[0] >> 4) == 4 {
		ipHeaderLen = int((msgBody.Data[0] & 0x0F) << 2)
	}

	dstPort := int(binary.BigEndian.Uint16(msgBody.Data[ipHeaderLen+2 : ipHeaderLen+4]))
	if dstPort < startingPort {
		return
	}

	pktNo := dstPort - startingPort
	i := pktNo % slots
	round := pktNo / slots

	var reqTime time.Time
	if value, ok := timestamps[i].Load(round); ok {
		reqTime = value.(time.Time)
	} else {
		return
	}

	rtt := float64(recvTime.Sub(reqTime).Nanoseconds()) / 1000000

	meta.MSamples[pktNo] = meta.RttSample{
		TTL:         getTTL(i),
		Round:       round + 1,
		ReplyIP:     replyIP,
		SendTime:    timeUtil.UnixPrecise(reqTime),
		RecvTime:    timeUtil.UnixPrecise(recvTime),
		RTT:         rtt,
		UdpDestPort: &dstPort,
	}

	if i == config.DirectHop && directHopIP == nil {
		directHopIP = replyIP
	}
}

func senderUDP(i int, dstIP net.IP) {
	defer close(senderDone[i])

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Println("[ping] [udp sender] error opening connection:", err)
		return
	}

	if dstIP.To4() == nil {
		err = ipv6.NewPacketConn(conn).SetHopLimit(getTTL(i))
	} else {
		err = ipv4.NewPacketConn(conn).SetTTL(getTTL(i))
	}
	if err != nil {
		log.Println("[ping] [udp sender] error setting ttl:", err)
		return
	}

	dstAddr := net.UDPAddr{
		IP:   dstIP,
		Port: startingPort + i - slots,
	}

	for r := 0; ; r++ {
		select {
		case <-channel.Stop:
			return
		case <-time.After(packetSendDelay):
			dstAddr.Port += slots
			timestamps[i].Store(r, time.Now())
			if _, err := conn.WriteTo(nil, &dstAddr); err != nil {
				log.Println("[ping] [udp sender] error sending packet:", err)
			}
		}
	}
}

func lostLoggerUDP(i int) (total, dropped int) {
	ttl := getTTL(i)

	timestamps[i].Range(func(key, value any) bool {
		r := key.(int)
		pktNo := i + r*slots
		total += 1

		if _, ok := meta.MSamples[pktNo]; !ok {
			udpPort := startingPort + pktNo
			meta.MSamples[pktNo] = meta.RttSample{
				TTL:         ttl,
				Round:       r + 1,
				SendTime:    timeUtil.UnixPrecise(value.(time.Time)),
				UdpDestPort: &udpPort,
			}
			dropped += 1
		}

		return true
	})

	return
}
