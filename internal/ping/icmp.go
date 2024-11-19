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
	"github.com/internet-equity/traceneck/internal/util"
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

	reqTime := timestamps[i][round]
	rtt := float64(recvTime.Sub(reqTime).Nanoseconds()) / 1000000

	meta.MSamples[pktNo] = meta.RttSample{
		TTL:       getTTL(i),
		Round:     round + 1,
		ReplyIP:   replyIP,
		SendTime:  util.GetTime(reqTime),
		RecvTime:  util.GetTime(recvTime),
		RTT:       rtt,
		IcmpSeqNo: pktNo,
	}

	if i == config.DirectHop && directHopIP == nil {
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

	timestamps[i] = make(map[int]time.Time)
	dstAddr := &net.IPAddr{IP: dstIP}

	for r := 0; ; r++ {
		select {
		case <-channel.Stop:
			return
		case <-time.After(packetSendDelay):
			msg := &icmp.Message{
				Type: typeEchoRequest,
				Code: 0,
				Body: &icmp.Echo{
					ID:   ID,
					Seq:  i + r*slots,
					Data: nil,
				},
			}

			if msgBytes, err := msg.Marshal(nil); err == nil {
				timestamps[i][r] = time.Now()
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

	for r, reqTime := range timestamps[i] {
		pktNo := i + r*slots
		total += 1

		if _, ok := meta.MSamples[pktNo]; !ok {
			meta.MSamples[pktNo] = meta.RttSample{
				TTL:       ttl,
				Round:     r + 1,
				SendTime:  util.GetTime(reqTime),
				IcmpSeqNo: pktNo,
			}

			dropped += 1
		}
	}

	return
}
