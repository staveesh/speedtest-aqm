package network

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"

	"github.com/internet-equity/traceneck/internal/channel"
	"github.com/internet-equity/traceneck/internal/config"
	"github.com/internet-equity/traceneck/internal/util"
)

// Ethernet Header: 14 bytes
// + IPv4 Header: 20 bytes (usually)
// OR IPv6 Header: 40 bytes
// + TCP Header: 20 bytes (usually)
// OR ICMP + UDP Header: 8 + 8 bytes
const captureSnapLen = 74

var (
	captureFilter string
	CapFile       string
)

func CaptureProcess() {
	defer close(channel.CaptureDone)

	if config.Tool == "ndt" {
		captureFilter = "port 80 or port 443"
	} else {
		captureFilter = "port 8080 or port 5060"
	}
	if config.PingType == "icmp" {
		captureFilter += " or icmp or icmp6"
	}
	if config.PingType == "udp" {
		captureFilter += " or (udp and outbound) or ((icmp or icmp6) and inbound)"
	}

	CapFile = util.GetFilePath(config.OutDir, "capture.pcap", config.Timestamp)
	capWriter, err := os.Create(CapFile)
	if err != nil {
		log.Println("[pcap] error opening pcap file:", err)
		return
	}
	defer capWriter.Close()

	if config.TShark {
		tsharkProcess()
	} else {
		pcapProcess(capWriter)
	}
}

func pcapProcess(capWriter *os.File) {
	handle, err := pcap.OpenLive(config.Interface, captureSnapLen, true, 100*time.Millisecond)
	if err != nil {
		log.Println("[pcap] error reading interface:", err)
		return
	}
	defer handle.Close()

	if err := handle.SetBPFFilter(captureFilter); err != nil {
		log.Println("[pcap] error setting bpf filter:", err)
		return
	}

	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()
	pcapWriter := pcapgo.NewWriter(capWriter)
	pcapWriter.WriteFileHeader(captureSnapLen, layers.LinkTypeEthernet)

	log.Println("[pcap] writing pcap to:", CapFile)
	defer log.Println("[pcap] stopped")

	for {
		select {
		case <-channel.Stop:
			return
		case packet := <-packets:
			if err := pcapWriter.WritePacket(
				packet.Metadata().CaptureInfo, packet.Data(),
			); err != nil {
				log.Println("[pcap] error writing to pcap:", err)
			}
		}
	}
}

func tsharkProcess() {
	Tshark := exec.Command(
		"tshark",
		"-F", "pcap",
		"-s", strconv.Itoa(captureSnapLen),
		"-i", config.Interface,
		"-f", captureFilter,
		"-w", CapFile)

	if err := Tshark.Start(); err != nil {
		log.Println("[tshark] error starting:", err)
		return
	}
	log.Println("[tshark] writing pcap to:", CapFile)

	<-channel.Stop

	if err := Tshark.Process.Kill(); err != nil {
		log.Println("[tshark] error stopping:", err)
	} else if _, err := Tshark.Process.Wait(); err != nil {
		log.Println("[tshark] error stopping:", err)
	} else {
		log.Println("[tshark] stopped")
	}
}
