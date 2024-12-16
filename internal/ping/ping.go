package ping

import (
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/meta"
	timeUtil "github.com/internet-equity/traceneck/internal/util/time"
)

const (
	icmpBufferSize = 56

	protocolICMP     = 1  // Internet Control Message
	protocolIPv6ICMP = 58 // ICMP for IPv6

	replyListenDelay = time.Second
	packetReadDelay  = 100 * time.Millisecond
	packetSendDelay  = 100 * time.Millisecond
)

var (
	slots      int
	timestamps []sync.Map

	stopListener channel.Type
	listenerDone channel.Type
	senderDone   []channel.Type

	timeExceededHandler func(net.IP, time.Time, *icmp.Message)
	sender              func(int, net.IP)
	lostLogger          func(int) (int, int)

	listenNetwork    string
	listenAddr       string
	msgProto         int
	typeEchoReply    icmp.Type
	typeTimeExceeded icmp.Type

	directHopIP net.IP
)

func getTTL(i int) (ttl int) {
	if i == 0 {
		ttl = config.DirectHop
	} else {
		ttl = i
	}

	return
}

func PingProcess() {
	defer close(channel.PingDone)

	slots = config.MaxTTL + 1
	timestamps = make([]sync.Map, slots)

	stopListener = make(channel.Type)
	listenerDone = make(channel.Type)
	senderDone = make([]channel.Type, slots)

	if config.PingType == "icmp" {
		timeExceededHandler = handleTimeExceededICMP
		sender = senderICMP
		lostLogger = lostLoggerICMP
	} else {
		timeExceededHandler = handleTimeExceededUDP
		sender = senderUDP
		lostLogger = lostLoggerUDP
	}

	for i := 0; i < slots; i++ {
		senderDone[i] = make(channel.Type)
	}

	<-channel.IPGrabbed

	if config.ServerIP.To4() == nil {
		listenNetwork = "ip6:ipv6-icmp"
		listenAddr = "::"
		msgProto = protocolIPv6ICMP
		typeEchoReply = ipv6.ICMPTypeEchoReply
		typeTimeExceeded = ipv6.ICMPTypeTimeExceeded
	} else {
		listenNetwork = "ip4:icmp"
		listenAddr = "0.0.0.0"
		msgProto = protocolICMP
		typeEchoReply = ipv4.ICMPTypeEchoReply
		typeTimeExceeded = ipv4.ICMPTypeTimeExceeded
	}

	log.Println("[ping] started")
	meta.MMeta.PingStartTime = timeUtil.UnixNow()

	go listener()

	for i := 1; i < slots; i++ {
		go sender(i, config.ServerIP)
	}

	if config.DirectHop == 0 {
		close(senderDone[0])
	} else {
		go pingDirect()
	}

	for i := 0; i < slots; i++ {
		<-senderDone[i]
	}

	time.Sleep(replyListenDelay)

	close(stopListener)
	<-listenerDone

	meta.MMeta.PingEndTime = timeUtil.UnixNow()
	log.Println("[ping] stopped")

	var (
		total   int
		dropped int
	)

	if directHopIP == nil {
		meta.MSamples[0] = meta.RttSample{
			TTL:       config.DirectHop,
			Round:     0,
			ReplyIP:   net.ParseIP("0.0.0.0"),
			IcmpSeqNo: new(int),
		}
	} else {
		total, dropped = lostLoggerICMP(0)
	}
	log.Println("[ping] hop:", config.DirectHop, "total:", total, "dropped:", dropped)

	for i := 1; i < slots; i++ {
		total, dropped = lostLogger(i)
		log.Println("[ping] hop:", i, "total:", total, "dropped:", dropped)
	}

	log.Println("[ping] logging complete")
}

func pingDirect() {
	for {
		select {
		case <-channel.Stop:
			close(senderDone[0])
			return
		case <-time.After(packetReadDelay):
			if directHopIP != nil {
				senderICMP(0, directHopIP)
				return
			}
		}
	}
}
