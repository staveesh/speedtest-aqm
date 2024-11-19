package ping

import (
	"log"
	"net"
	"time"

	"golang.org/x/net/icmp"
)

func listener() {
	defer close(listenerDone)

	conn, err := icmp.ListenPacket(listenNetwork, listenAddr)
	if err != nil {
		log.Println("[ping] [listener] error opening connection:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, icmpBufferSize)

	for {
		select {
		case <-stopListener:
			return
		default:
			if err := conn.SetReadDeadline(time.Now().Add(packetReadDelay)); err != nil {
				log.Println("[ping] [listener] error setting deadline:", err)
				return
			}

			n, peer, err := conn.ReadFrom(buffer)
			if err != nil {
				break
			}
			recvTime := time.Now()

			msg, err := icmp.ParseMessage(msgProto, buffer[:n])
			if err != nil {
				break
			}

			switch msg.Type {
			case typeEchoReply:
				handleEchoReply(net.ParseIP(peer.String()), recvTime, msg)
			case typeTimeExceeded:
				timeExceededHandler(net.ParseIP(peer.String()), recvTime, msg)
			}
		}
	}
}
