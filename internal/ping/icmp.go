package ping

import (
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

var ID = os.Getpid() & 0xffff

func handleEchoReply(replyIP net.IP, recvTime time.Time, msg *icmp.Message) {
	msgBody, ok := msg.Body.(*icmp.Echo)
	if !ok || msgBody.ID != ID || msgBody.Seq < 0 {
		return
	}

	pktNo := msgBody.Seq
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
		TTL:       getTTL(i),
		Round:     round + 1,
		ReplyIP:   replyIP,
		SendTime:  timeUtil.UnixPrecise(reqTime),
		RecvTime:  timeUtil.UnixPrecise(recvTime),
		RTT:       rtt,
		IcmpSeqNo: pktNo,
	}

	if directHopIP == nil && i == config.DirectHop {
		directHopIP = replyIP
	}
}

func handleTimeExceededICMP(replyIP net.IP, recvTime time.Time, msg *icmp.Message) {
	msgBody, ok := msg.Body.(*icmp.TimeExceeded)
	if !ok {
		return
	}

	reqMsg, err := icmp.ParseMessage(msgProto, msgBody.Data[20:])
	if err != nil {
		return
	}

	handleEchoReply(replyIP, recvTime, reqMsg)
}

func senderICMP(i int, dstIP net.IP) {
	defer close(senderDone[i])

	conn, err := icmp.ListenPacket(listenNetwork, listenAddr)
	if err != nil {
		log.Println("[ping] [icmp sender] error opening connection:", err)
	}
	defer conn.Close()

	var typeEchoRequest icmp.Type
	if dstIP.To4() == nil {
		err = conn.IPv6PacketConn().SetHopLimit(getTTL(i))
		typeEchoRequest = ipv6.ICMPTypeEchoRequest
	} else {
		err = conn.IPv4PacketConn().SetTTL(getTTL(i))
		typeEchoRequest = ipv4.ICMPTypeEcho
	}
	if err != nil {
		log.Println("[ping] [icmp sender] error setting ttl:", err)
		return
	}

	dstAddr := &net.IPAddr{IP: dstIP}
	msg := &icmp.Message{
		Type: typeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:  ID,
			Seq: i - slots,
		},
	}

	for r := 0; ; r++ {
		select {
		case <-channel.Stop:
			return
		case <-time.After(packetSendDelay):
			msg.Body.(*icmp.Echo).Seq += slots
			if msgBytes, err := msg.Marshal(nil); err == nil {
				timestamps[i].Store(r, time.Now())
				if _, err := conn.WriteTo(msgBytes, dstAddr); err != nil {
					log.Println("[ping] [icmp sender] error sending packet:", err)
				}
			} else {
				log.Println("[ping] [icmp sender] error encoding packet:", err)
			}
		}
	}
}

func lostLoggerICMP(i int) (total, dropped int) {
	ttl := getTTL(i)

	timestamps[i].Range(func(key, value any) bool {
		r := key.(int)
		pktNo := i + r*slots
		total += 1

		if _, ok := meta.MSamples[pktNo]; !ok {
			meta.MSamples[pktNo] = meta.RttSample{
				TTL:       ttl,
				Round:     r + 1,
				SendTime:  timeUtil.UnixPrecise(value.(time.Time)),
				IcmpSeqNo: pktNo,
			}
			dropped += 1
		}

		return true
	})

	return
}
